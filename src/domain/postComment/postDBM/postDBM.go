package postDBM

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/types/UITypes"

	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

func New(ctx context.Context, tx pgx.Tx, clientUsername string, mediaCloudNames []string, postType, description string, at int64, mentions, hashtags []string) (post UITypes.NewPost, err error) {
	newPost, err := pgDB.QueryRowTypeTx[UITypes.NewPost](
		ctx, tx,
		/* sql */ `
		SELECT * FROM new_post($1, $2, $3, $4, $5, $6, $7)
		`, clientUsername, postType, mediaCloudNames, description, at, mentions, hashtags,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewPost{}, fiber.ErrInternalServerError
	}

	return *newPost, nil
}

func Get(ctx context.Context, clientUsername, postId string) (UITypes.Post, error) {
	post, err := pgDB.QueryRowType[UITypes.Post](
		ctx,
		/* sql */ `
		SELECT * FROM get_post($1, $2)
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.Post{}, fiber.ErrInternalServerError
	}

	return *post, nil
}

func Delete(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		WITH del AS (
			UPDATE posts
			SET deleted = true, deleted_at = now()
			WHERE id_ = $1 AND owner_user = $2
			RETURNING true AS done
		)
		UPDATE users
		SET posts_count = posts_count - 1
		WHERE username = $2 AND (SELECT done FROM del) = true
		RETURNING true
		`, postId, clientUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func ReactTo(ctx context.Context, tx pgx.Tx, clientUsername, postId, emoji string, at int64) (string, error) {
	reactNotifId, err := pgDB.QueryRowFieldTx[string](
		ctx, tx,
		/* sql */ `
		SELECT * FROM react_to_post($1, $2, $3, $4)
		`, clientUsername, postId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return "", helpers.HandleDBError(err)
	}

	return *reactNotifId, nil
}

func GetReactors(ctx context.Context, postId string, limit int64, cursor float64) ([]*UITypes.ReactorSnippet, error) {
	reactors, err := pgDB.QueryRowsType[UITypes.ReactorSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_post_reactors($1, $2, $3)
		`, postId, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return reactors, nil
}

func RemoveReaction(ctx context.Context, tx pgx.Tx, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		SELECT * FROM unreact_to_post($1, $2)
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func CommentOn(ctx context.Context, tx pgx.Tx, clientUsername, postId, commentText, attachmentCloudName string, at int64, mentions []string) (UITypes.NewComment, error) {
	newComment, err := pgDB.QueryRowTypeTx[UITypes.NewComment](
		ctx, tx,
		/* sql */ `
		SELECT * FROM comment_on_post($1, $2, $3, $4, $5, $6)
		`, clientUsername, postId, commentText, attachmentCloudName, at, mentions,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, helpers.HandleDBError(err)
	}

	return *newComment, nil
}

func GetComments(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) ([]*UITypes.Comment, error) {
	comments, err := pgDB.QueryRowsType[UITypes.Comment](
		ctx,
		/* sql */ `
		SELECT * FROM get_post_comments($1, $2, $3, $4)
		`, clientUsername, postId, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return comments, nil
}

func RemoveComment(ctx context.Context, tx pgx.Tx, clientUsername, postId, commentId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		SELECT * FROM uncomment_on_post($1, $2, $3)
		`, clientUsername, postId, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

type RepostT struct {
	Id      string `db:"repost_id"`
	Cursor  int64  `db:"repost_cursor"`
	NotifId string `db:"repost_notif_id"`
}

func Repost(ctx context.Context, tx pgx.Tx, clientUsername, postId string) (RepostT, error) {
	repost, err := pgDB.QueryRowType[RepostT](
		ctx,
		/* sql */ `
		SELECT * FROM repost($1, $2)
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return RepostT{}, fiber.ErrInternalServerError
	}

	return *repost, nil
}

func Save(ctx context.Context, tx pgx.Tx, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		INSERT INTO post_saves(username, post_id)
		VALUES ($1, $2)
		RETURNING true
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, helpers.HandleDBError(err)
	}

	return *done, nil
}

func Unsave(ctx context.Context, tx pgx.Tx, clientUsername, postId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		DELETE FROM post_saves
		WHERE username = $1 AND post_id = $2
		RETURNING true
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}
