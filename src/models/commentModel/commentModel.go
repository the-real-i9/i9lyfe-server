package commentModel

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/models/modelHelpers"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

func Get(ctx context.Context, clientUsername, commentId string) (comment UITypes.Comment, err error) {
	comment, err = modelHelpers.BuildCommentUIFromCache(ctx, commentId, clientUsername)
	if err != nil {
		helpers.LogError(err)
		return UITypes.Comment{}, fiber.ErrInternalServerError
	}

	return comment, nil
}

type reactToT struct {
	CommentOwnerUser string `db:"owner_user"`
	RxnCursor        int64  `db:"rxn_cursor"`
}

func ReactTo(ctx context.Context, clientUsername, commentId, emoji string, at int64) (reactToT, error) {
	res, err := pgDB.QueryRowType[reactToT](
		ctx,
		/* sql */ `
		WITH react_to AS (
			INSERT INTO user_reacts_to_comment(username, comment_id, emoji, at_)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT no_dup_comment_rxn DO UPDATE 
			SET emoji = $3, at_ = $4
			RETURNING cursor_
		)
		SELECT c.username AS owner_user, r.cursor_ AS rxn_cursor 
		FROM user_comments_on c, react_to r
		WHERE comment_id = $2 AND r.cursor_ IS NOT NULL
		`, clientUsername, commentId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return reactToT{}, helpers.HandleDBError(err)
	}

	return *res, nil
}

func GetReactors(ctx context.Context, clientUsername, commentId string, limit int, cursor float64) (reactors []UITypes.ReactorSnippet, err error) {
	reactorMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("reacted_comment:%s:reactors", commentId), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	reactors, err = modelHelpers.ReactorMembersForUIReactorSnippets(ctx, reactorMembers, "comment", commentId)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return reactors, nil
}

/* func GetReactorsWithReaction(ctx context.Context, clientUsername, commentId, reaction string, limit int, offset time.Time) ([]any, error) {

} */

func RemoveReaction(ctx context.Context, clientUsername, commentId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
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

type NewCommentT struct {
	Id                 string `json:"id" db:"comment_id"`
	OwnerUser          any    `json:"owner_user" db:"owner_user"`
	CommentText        string `json:"comment_text" db:"comment_text"`
	AttachmentUrl      string `json:"attachment_url" db:"attachment_url"`
	At                 int64  `json:"at" db:"at_"`
	Cursor             int64  `json:"cursor" db:"cursor_"`
	ParentCommentOwner string `json:"-" db:"parent_comment_owner"`
}

func CommentOn(ctx context.Context, clientUsername, parentCommentId, commentText, attachmentCloudName string, at int64) (NewCommentT, error) {
	newComment, err := pgDB.QueryRowType[NewCommentT](
		ctx,
		/* sql */ `
		WITH comment_on AS (
			INSERT INTO user_comments_on(username, parent_comment_id, comment_text, attachment_url, at_)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING comment_id, username AS owner_user, comment_text, attachment_url, at_
		)
		SELECT comment_id, owner_user, comment_text, attachment_url, at_, cursor_, (SELECT username FROM user_comments_on WHERE comment_id = $2) AS parent_comment_owner FROM comment_on
		`, clientUsername, parentCommentId, commentText, attachmentCloudName, at,
	)
	if err != nil {
		helpers.LogError(err)
		return NewCommentT{}, helpers.HandleDBError(err)
	}

	return *newComment, nil
}

func CommentOnExtras(ctx context.Context, newCommentId string, mentions []string) error {
	var err error

	tx, err := appGlobals.DBPool.Begin(ctx)
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

func GetComments(ctx context.Context, clientUsername, commentId string, limit int, cursor float64) (comments []UITypes.Comment, err error) {
	commentMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("commented_comment:%s:comments", commentId), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	comments, err = modelHelpers.CommentMembersForUIComments(ctx, commentMembers, clientUsername)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return comments, nil
}

func RemoveComment(ctx context.Context, clientUsername, parentCommentId, commentId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
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
