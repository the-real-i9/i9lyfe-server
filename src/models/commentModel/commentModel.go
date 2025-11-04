package commentModel

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var dbPool = appGlobals.DBPool

func Get(ctx context.Context, clientUsername, commentId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (comment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(ownerUser:User), (clientUser:User{ username: $client_username })
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_COMMENT]->(comment)
		WITH comment, 
			toString(comment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		RETURN comment { .*, owner_user, created_at, client_reaction } AS found_comment
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundComment, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_comment")

	return foundComment, nil
}

func ReactTo(ctx context.Context, clientUsername, commentId, emoji string, at int64) (string, error) {
	commentOwner, err := db.QueryRowField[string](
		ctx,
		/* sql */ `
		WITH react_to AS (
			INSERT INTO user_reacts_to_comment(username, comment_id, emoji, at_)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT no_dup_comment_rxn DO UPDATE 
			SET emoji = $3, at_ = $4
			RETURNING true AS done
		)
		SELECT username FROM user_comments_on
		WHERE comment_id = $2 AND (SELECT done FROM react_to) = true
		`, clientUsername, commentId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	return *commentOwner, nil
}

func GetReactors(ctx context.Context, clientUsername, commentId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT]-(reactor:User)
		WHERE rxn.at < $offset
		OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
		WITH reactor, 
			rxn, 
			CASE fur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY rxn.at DESC
		LIMIT $limit
		RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors")

	return reactors, nil
}

func GetReactorsWithReaction(ctx context.Context, clientUsername, commentId, reaction string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT { reaction: $reaction }]-(reactor:User)
		WHERE rxn.at < $offset
		OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
		WITH reactor, 
			rxn, 
			CASE fur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY rxn.at DESC
		LIMIT $limit
		RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_wrxn
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
			"reaction":        reaction,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors_wrxn")

	return reactors, nil
}

func RemoveReaction(ctx context.Context, clientUsername, commentId string) (bool, error) {
	done, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_reacts_to_comment
		WHERE username = $1 AND comment_id = $2
		RETURNING true AS done
		`, clientUsername, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

type newCommentT struct {
	Id                 string `json:"id" db:"comment_id"`
	CommentText        string `json:"comment_text" db:"comment_text"`
	AttachmentUrl      string `json:"attachment_url" db:"attachment_url"`
	At                 int64  `json:"at" db:"at_"`
	ParentCommentOwner string `json:"-" db:"parent_comment_owner"`
}

func CommentOn(ctx context.Context, clientUsername, parentCommentId, commentText, attachmentUrl string, at int64) (newCommentT, error) {
	newComment, err := db.QueryRowType[newCommentT](
		ctx,
		/* sql */ `
		WITH comment_on AS (
			INSERT INTO user_comments_on(username, parent_comment_id, comment_text, attachment_url, at_)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING comment_id, comment_text, attachment_url, at_
		)
		SELECT comment_id, comment_text, attachment_url, at_, (SELECT username FROM user_comments_on WHERE parent_comment_id = $2) AS parent_comment_owner FROM comment_on
		`, clientUsername, parentCommentId, commentText, attachmentUrl, at,
	)
	if err != nil {
		helpers.LogError(err)
		return newCommentT{}, fiber.ErrInternalServerError
	}

	return *newComment, nil
}

func CommentOnExtras(ctx context.Context, newCommentId string, mentions []string) error {
	var err error

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	for _, mu := range mentions {
		_, err := tx.Exec(
			ctx,
			/* sql */ `
				INSERT INTO comment_mentions_user (comment_id, username)
				VALUES ($1, $2)
				`, newCommentId, mu,
		)
		if err != nil {
			helpers.LogError(err)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func GetComments(ctx context.Context, clientUsername, commentId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (parentComment:Comment{ id: $comment_id })<-[:COMMENT_ON_COMMENT]-(childComment:Comment WHERE childComment.created_at < $offset)<-[:WRITES_COMMENT]-(ownerUser:User)

		OPTIONAL MATCH (childComment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })
		WITH childComment, 
			toString(childComment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		ORDER BY childComment.created_at DESC, childComment.reactions_count DESC, childComment.comments_count DESC
		LIMIT $limit
		RETURN collect(childComment {.*, created_at, owner_user, client_reaction }) AS comments
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	comments, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "comments")

	return comments, nil
}

func RemoveComment(ctx context.Context, clientUsername, parentCommentId, commentId string) (bool, error) {
	done, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_comments_on
		WHERE username = $1 AND parent_comment_id = $2 AND comment_id = $3
		RETURNING true AS done
		`, clientUsername, parentCommentId, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}
