package commentDBM

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

func Get(ctx context.Context, clientUsername, commentId string) (UITypes.Comment, error) {
	comment, err := pgDB.QueryRowType[UITypes.Comment](
		ctx,
		/* sql */ `
		SELECT * FROM get_comment($1, $2)
		`, clientUsername, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.Comment{}, fiber.ErrInternalServerError
	}

	return *comment, nil
}

func ReactTo(ctx context.Context, tx pgx.Tx, clientUsername, commentId, emoji string, at int64) (string, error) {
	reactNotifId, err := pgDB.QueryRowFieldTx[string](
		ctx, tx,
		/* sql */ `
		SELECT * FROM react_to_comment($1, $2, $3, $4)
		`, clientUsername, commentId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return "", helpers.HandleDBError(err)
	}

	return *reactNotifId, nil
}

func GetReactors(ctx context.Context, commentId string, limit int64, cursor float64) ([]*UITypes.ReactorSnippet, error) {
	reactors, err := pgDB.QueryRowsType[UITypes.ReactorSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_comment_reactors($1, $2, $3)
		`, commentId, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return reactors, nil
}

/* func GetReactorsWithReaction(ctx context.Context, clientUsername, commentId, reaction string, limit int64, offset time.Time) ([]any, error) {

} */

func RemoveReaction(ctx context.Context, tx pgx.Tx, clientUsername, commentId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		SELECT * FROM unreact_to_comment($1, $2)
		`, clientUsername, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func CommentOn(ctx context.Context, tx pgx.Tx, clientUsername, parentCommentId, commentText, attachmentCloudName string, at int64, mentions []string) (UITypes.NewComment, error) {
	newComment, err := pgDB.QueryRowTypeTx[UITypes.NewComment](
		ctx, tx,
		/* sql */ `
		SELECT * FROM comment_on_comment($1, $2, $3, $4, $5, $6)
		`, clientUsername, parentCommentId, commentText, attachmentCloudName, at, mentions,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, helpers.HandleDBError(err)
	}

	return *newComment, nil
}

func GetComments(ctx context.Context, clientUsername, commentId string, limit int64, cursor float64) ([]*UITypes.Comment, error) {
	comments, err := pgDB.QueryRowsType[UITypes.Comment](
		ctx,
		/* sql */ `
		SELECT * FROM get_comment_comments($1, $2, $3, $4)
		`, clientUsername, commentId, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return comments, nil
}

func RemoveComment(ctx context.Context, tx pgx.Tx, clientUsername, parentCommentId, commentId string) (bool, error) {
	done, err := pgDB.QueryRowFieldTx[bool](
		ctx, tx,
		/* sql */ `
		SELECT * FROM uncomment_on_comment($1, $2, $3)
		`, clientUsername, parentCommentId, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}
