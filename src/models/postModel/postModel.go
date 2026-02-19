package postModel

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/models/modelHelpers"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

type NewPostT struct {
	Id          string   `msgpack:"id" db:"id_"`
	Type        string   `msgpack:"type" db:"type_"`
	MediaUrls   []string `msgpack:"media_urls" db:"media_urls"`
	Description string   `msgpack:"description"`
	CreatedAt   int64    `msgpack:"created_at" db:"created_at"`
	OwnerUser   any      `msgpack:"owner_user" db:"owner_user"`
	Cursor      int64    `msgpack:"cursor" db:"cursor_"`
}

func New(ctx context.Context, clientUsername string, mediaCloudNames []string, postType, description string, at int64) (post NewPostT, err error) {
	newPost, err := pgDB.QueryRowType[NewPostT](
		ctx,
		/* sql */ `
		INSERT INTO posts (owner_user, type_, media_urls, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id_, owner_user, type_, media_urls, description, created_at, cursor_
		`, clientUsername, postType, mediaCloudNames, description, at,
	)
	if err != nil {
		helpers.LogError(err)
		return NewPostT{}, fiber.ErrInternalServerError
	}

	newPost.Cursor += time.Now().UnixMicro()

	return *newPost, nil
}

func NewPostExtras(ctx context.Context, newPostId string, mentions, hashtags []string) error {
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
				INSERT INTO post_mentions_user (post_id, username)
				VALUES ($1, $2)
				ON CONFLICT ON CONSTRAINT no_dup_post_ment DO NOTHING
				`, newPostId, mu,
		)
		if err != nil {
			helpers.LogError(err)
			if !errors.Is(err, &pgconn.PgError{}) {
				return err
			}
		}
	}

	for _, ht := range hashtags {
		_, err := tx.Exec(
			ctx,
			/* sql */ `
				INSERT INTO hashtags (htname)
				VALUES ($1)
				ON CONFLICT ON CONSTRAINT hashtags_pkey DO NOTHING
				`, ht,
		)
		if err != nil {
			helpers.LogError(err)
			return err
		}

		_, err = tx.Exec(
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

func Get(ctx context.Context, clientUsername, postId string) (post UITypes.Post, err error) {
	post, err = modelHelpers.BuildPostUIFromCache(ctx, postId, clientUsername)
	if err != nil {
		helpers.LogError(err)
		return UITypes.Post{}, fiber.ErrInternalServerError
	}

	return post, nil
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

type reactToT struct {
	PostOwnerUser string `db:"owner_user"`
	RxnCursor     int64  `db:"rxn_cursor"`
}

func ReactTo(ctx context.Context, clientUsername, postId, emoji string, at int64) (reactToT, error) {
	res, err := pgDB.QueryRowType[reactToT](
		ctx,
		/* sql */ `
		WITH react_to AS (
			INSERT INTO user_reacts_to_post(username, post_id, emoji, at_)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT no_dup_post_rxn DO UPDATE 
			SET emoji = $3, at_ = $4
			RETURNING cursor_
		)
		SELECT p.owner_user AS owner_user, r.cursor_ AS rxn_cursor 
		FROM posts p, react_to r
		WHERE id_ = $2 AND r.cursor_ IS NOT NULL
		`, clientUsername, postId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return reactToT{}, helpers.HandleDBError(err)
	}

	return *res, nil
}

func GetReactors(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) (reactors []UITypes.ReactorSnippet, err error) {
	reactorMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("reacted_post:%s:reactors", postId), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: limit,
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	reactors, err = modelHelpers.ReactorMembersForUIReactorSnippets(ctx, reactorMembers, "post", postId)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return reactors, nil
}

/* func GetReactorsWithReaction(ctx context.Context, clientUsername, postId, reaction string, limit int64, offset time.Time) ([]any, error) {

} */

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

type NewCommentT struct {
	Id            string `msgpack:"id" db:"comment_id"`
	OwnerUser     any    `msgpack:"owner_user" db:"owner_user"`
	CommentText   string `msgpack:"comment_text" db:"comment_text"`
	AttachmentUrl string `msgpack:"attachment_url" db:"attachment_url"`
	At            int64  `msgpack:"at" db:"at_"`
	Cursor        int64  `msgpack:"cursor" db:"cursor_"`
	PostOwner     string `msgpack:"-" db:"post_owner"`
}

func CommentOn(ctx context.Context, clientUsername, postId, commentText, attachmentCloudName string, at int64) (NewCommentT, error) {
	newComment, err := pgDB.QueryRowType[NewCommentT](
		ctx,
		/* sql */ `
		WITH comment_on AS (
			INSERT INTO user_comments_on(username, post_id, comment_text, attachment_url, at_)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING comment_id, username AS owner_user, comment_text, attachment_url, at_
		)
		SELECT comment_id, owner_user, comment_text, attachment_url, at_, cursor_, (SELECT p.owner_user FROM posts p WHERE id_ = $2) AS post_owner FROM comment_on
		`, clientUsername, postId, commentText, attachmentCloudName, at,
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
			if !errors.Is(err, &pgconn.PgError{}) {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func GetComments(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) (comments []UITypes.Comment, err error) {
	commentMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("commented_post:%s:comments", postId), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: limit,
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

type RepostT struct {
	Id             string   `msgpack:"id" db:"id_"`
	Type           string   `msgpack:"type" db:"type_"`
	MediaUrls      []string `msgpack:"media_urls" db:"media_urls"`
	Description    string   `msgpack:"description"`
	CreatedAt      int64    `msgpack:"created_at" db:"created_at"`
	OwnerUser      any      `msgpack:"owner_user" db:"owner_user"`
	ReposterUser   any      `msgpack:"reposter_user" db:"reposted_by_user"`
	RepostedPostId string   `msgpack:"reposted_post_id" db:"reposted_post_id"`
	Cursor         int64    `msgpack:"cursor" db:"cursor_"`
}

func Repost(ctx context.Context, clientUsername, postId string, at int64) (RepostT, error) {
	repost, err := pgDB.QueryRowType[RepostT](
		ctx,
		/* sql */ `
		INSERT INTO posts (owner_user, type_, media_urls, description, created_at, reposted_by_user, reposted_post_id)
		SELECT owner_user, type_, media_urls, description, $3, $1, $2 FROM posts
		WHERE id_ = $2
		RETURNING id_, owner_user, type_, media_urls, description, created_at, reposted_by_user, reposted_post_id, cursor_
		`, clientUsername, postId, at,
	)
	if err != nil {
		helpers.LogError(err)
		return RepostT{}, fiber.ErrInternalServerError
	}

	repost.Cursor += time.Now().UnixMicro()

	return *repost, nil
}

func Save(ctx context.Context, clientUsername, postId string) (int64, error) {
	saveCursor, err := pgDB.QueryRowField[int64](
		ctx,
		/* sql */ `
		INSERT INTO user_saves_post(username, post_id)
		VALUES ($1, $2)
		RETURNING cursor_
		`, clientUsername, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return 0, helpers.HandleDBError(err)
	}

	return *saveCursor, nil
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
