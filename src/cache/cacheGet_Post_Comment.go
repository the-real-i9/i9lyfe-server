package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cloudStorageService"

	"github.com/redis/go-redis/v9"
)

func GetPost[T any](ctx context.Context, postId string) (post T, err error) {
	postJson, err := rdb().HGet(ctx, "posts", postId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return post, err
	}

	postMap := helpers.FromJson[map[string]any](postJson)

	if err := cloudStorageService.PostMediaCloudNamesToUrl(postMap); err != nil {
		return post, err
	}

	return helpers.MapToStruct[T](postMap), nil
}

func GetComment[T any](ctx context.Context, commentId string) (comment T, err error) {
	commentJson, err := rdb().HGet(ctx, "comments", commentId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return comment, err
	}

	commentMap := helpers.FromJson[map[string]any](commentJson)

	if err := cloudStorageService.CommentAttachCloudNameToUrl(commentMap); err != nil {
		return comment, err
	}

	return helpers.MapToStruct[T](commentMap), nil
}

func GetPostReactionsCount(ctx context.Context, postId string) (int64, error) {
	count, err := rdb().HLen(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetPostCommentsCount(ctx context.Context, postId string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("commented_post:%s:comments", postId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetPostSavesCount(ctx context.Context, postId string) (int64, error) {
	count, err := rdb().SCard(ctx, fmt.Sprintf("saved_post:%s:saves", postId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetPostRepostsCount(ctx context.Context, postId string) (int64, error) {
	count, err := rdb().SCard(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetCommentReactionsCount(ctx context.Context, commentId string) (int64, error) {
	count, err := rdb().HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetCommentCommentsCount(ctx context.Context, parentCommentId string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("commented_comment:%s:comments", parentCommentId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}
