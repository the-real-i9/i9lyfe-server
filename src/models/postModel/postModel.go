package postModel

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
			return err
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
		return UITypes.Post{}, err
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

func GetReactors(ctx context.Context, clientUsername, postId string, limit int, cursor float64) (reactors []UITypes.ReactorSnippet, err error) {
	reactorMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("reacted_post:%s:reactors", postId), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
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

/* func GetReactorsWithReaction(ctx context.Context, clientUsername, postId, reaction string, limit int, offset time.Time) ([]any, error) {

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

type newCommentT struct {
	Id            string `json:"id" db:"comment_id"`
	OwnerUser     string `json:"owner_user" db:"owner_user"`
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
			RETURNING comment_id, username AS owner_user, comment_text, attachment_url, at_
		)
		SELECT comment_id, owner_user, comment_text, attachment_url, at_, (SELECT p.owner_user FROM posts p WHERE id_ = $2) AS post_owner FROM comment_on
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

func GetComments(ctx context.Context, clientUsername, postId string, limit int, cursor float64) (comments []UITypes.Comment, err error) {
	commentMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("commented_post:%s:comments", postId), &redis.ZRangeBy{
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
