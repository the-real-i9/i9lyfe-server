package postModel

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func dbPool() *pgxpool.Pool {
	return appGlobals.DBPool
}

type newPostT struct {
	Id          string   `json:"id" db:"id_"`
	Type        string   `json:"type" db:"type_"`
	MediaUrls   []string `json:"media_urls" db:"media_urls"`
	Description string   `json:"description"`
	CreatedAt   int64    `json:"created_at" db:"created_at"`
	OwnerUser   any      `json:"owner_user" db:"owner_user"`
}

func New(ctx context.Context, clientUsername string, mediaUrls []string, postType, description string, at int64) (post newPostT, err error) {
	newPost, err := pgDB.QueryRowType[newPostT](
		ctx,
		/* sql */ `
		INSERT INTO posts (owner_user, type_, media_urls, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id_, owner_user, type_, media_urls, description, created_at
		`, clientUsername, postType, mediaUrls, description, at,
	)
	if err != nil {
		helpers.LogError(err)
		return newPostT{}, fiber.ErrInternalServerError
	}

	return *newPost, nil
}

func NewPostExtras(ctx context.Context, newPostId string, mentions, hashtags []string) error {
	var err error

	tx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	for _, mu := range mentions {
		_, err := tx.Exec(
			ctx,
			/* sql */ `
				INSERT INTO post_mentions_user (post_id, username)
				VALUES ($1, $2)
				ON CONFLICT ON CONSTRAINT no_dup_post_ment DO NOTHING
				`, newPostId, mu,
		)
		if err != nil {
			helpers.LogError(err)
			return err
		}
	}

	for _, ht := range hashtags {
		_, err := tx.Exec(
			ctx,
			/* sql */ `
				INSERT INTO post_includes_hashtag (post_id, htname)
				VALUES ($1, $2)
				ON CONFLICT ON CONSTRAINT no_dup_htname DO NOTHING
				`, newPostId, ht,
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

func Get(ctx context.Context, clientUsername, postId string) (map[string]any, error) {
	res, err := pgDB.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser:User), (clientUser:User{ username: $client_username })

		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
		OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		
		WITH post, 
			toString(post.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction, 
			CASE csaves 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_saved, 
			CASE creposts 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_reposted
		RETURN post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted } AS found_post
    `,
		map[string]any{
			"post_id":         postId,
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

	foundPost, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_post")

	return foundPost, nil
}

func Delete(ctx context.Context, clientUsername, postId string) (mentionedUsers []string, err error) {
	mentionedUsersPs, err := pgDB.QueryRowsField[string](
		ctx,
		/* sql */ `
		WITH delete_post AS (
			UPDATE posts
			SET deleted = true, deleted_at = now()
			WHERE id_ = $1 AND owner_user = $2
			RETURNING true AS done
		)
		SELECT username FROM post_mentions_user
		WHERE post_id = $1 AND (SELECT done FROM delete_post) = true
		`, postId, clientUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	for _, s := range mentionedUsersPs {
		mentionedUsers = append(mentionedUsers, *s)
	}

	return mentionedUsers, nil
}

func ReactTo(ctx context.Context, clientUsername, postId, emoji string, at int64) (string, error) {
	postOwnerUser, err := pgDB.QueryRowField[string](
		ctx,
		/* sql */ `
		WITH react_to AS (
			INSERT INTO user_reacts_to_post(username, post_id, emoji, at_)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT no_dup_post_rxn DO UPDATE 
			SET emoji = $3, at_ = $4
			RETURNING true AS done
		)
		SELECT owner_user FROM posts
		WHERE id_ = $2 AND (SELECT done FROM react_to) = true
		`, clientUsername, postId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	return *postOwnerUser, nil
}

func GetReactors(ctx context.Context, clientUsername, postId string, limit int, offset time.Time) ([]any, error) {
	res, err := pgDB.Query(
		ctx,
		`
		MATCH (:Post{ id: $post_id })<-[rxn:REACTS_TO_POST]-(reactor:User)
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
			"post_id":         postId,
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

func GetReactorsWithReaction(ctx context.Context, clientUsername, postId, reaction string, limit int, offset time.Time) ([]any, error) {
	res, err := pgDB.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[rxn:REACTS_TO_POST { reaction: toString($reaction) }]-(reactor:User)
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
			"post_id":         postId,
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

func RemoveReaction(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_reacts_to_post
		WHERE username = $1 AND post_id = $2
		RETURNING true AS done
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

type newCommentT struct {
	Id            string `json:"id" db:"comment_id"`
	CommentText   string `json:"comment_text" db:"comment_text"`
	AttachmentUrl string `json:"attachment_url" db:"attachment_url"`
	At            int64  `json:"at" db:"at_"`
	PostOwner     string `json:"-" db:"post_owner"`
}

func CommentOn(ctx context.Context, clientUsername, postId, commentText, attachmentUrl string, at int64) (newCommentT, error) {
	newComment, err := pgDB.QueryRowType[newCommentT](
		ctx,
		/* sql */ `
		WITH comment_on AS (
			INSERT INTO user_comments_on(username, post_id, comment_text, attachment_url, at_)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING comment_id, comment_text, attachment_url, at_
		)
		SELECT comment_id, comment_text, attachment_url, at_, (SELECT owner_user FROM posts WHERE id_ = $2) AS post_owner FROM comment_on
		`, clientUsername, postId, commentText, attachmentUrl, at,
	)
	if err != nil {
		helpers.LogError(err)
		return newCommentT{}, fiber.ErrInternalServerError
	}

	return *newComment, nil
}

func CommentOnExtras(ctx context.Context, newCommentId string, mentions []string) error {
	var err error

	tx, err := dbPool().Begin(ctx)
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
				ON CONFLICT ON CONSTRAINT no_dup_comment_ment DO NOTHING
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

func GetComments(ctx context.Context, clientUsername, postId string, limit int, offset time.Time) ([]any, error) {
	res, err := pgDB.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[:COMMENT_ON_POST]-(comment:Comment WHERE comment.created_at < $offset)<-[:WRITES_COMMENT]-(ownerUser:User)

		OPTIONAL MATCH (comment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })

		WITH comment, 
			toString(comment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		ORDER BY comment.created_at DESC
		LIMIT $limit
		RETURN collect(comment {.*, owner_user, created_at, client_reaction }) AS comments
		`,
		map[string]any{
			"post_id":         postId,
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

func RemoveComment(ctx context.Context, clientUsername, postId, commentId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_comments_on
		WHERE username = $1 AND post_id = $2 AND comment_id = $3
		RETURNING true AS done
		`, clientUsername, postId, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

type repostT struct {
	Id             string   `json:"id" db:"id_"`
	Type           string   `json:"type" db:"type_"`
	MediaUrls      []string `json:"media_urls" db:"media_urls"`
	Description    string   `json:"description"`
	CreatedAt      int64    `json:"created_at" db:"created_at"`
	OwnerUser      string   `json:"owner_user" db:"owner_user"`
	ReposterByUser string   `json:"reposted_by_user" db:"reposted_by_user"`
}

func Repost(ctx context.Context, clientUsername, postId string, at int64) (repostT, error) {
	repost, err := pgDB.QueryRowType[repostT](
		ctx,
		/* sql */ `
		INSERT INTO posts (owner_user, type_, media_urls, description, created_at, reposted_by_user)
		SELECT owner_user, type_, media_urls, description, $3, $1 FROM posts
		WHERE id_ = $2
		RETURNING id_, owner_user, type_, media_urls, description, created_at, reposted_by_user
		`, clientUsername, postId, at,
	)
	if err != nil {
		helpers.LogError(err)
		return repostT{}, fiber.ErrInternalServerError
	}

	return *repost, nil
}

func Save(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		INSERT INTO user_saves_post(username, post_id)
		VALUES ($1, $2)
		RETURNING true AS done
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func Unsave(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_saves_post
		WHERE username = $1 AND post_id = $2
		RETURNING true AS done
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}
