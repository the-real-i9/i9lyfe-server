package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

func GetPost[T any](ctx context.Context, postId string) (post T, err error) {
	postJson, err := rdb().HGet(ctx, "posts", postId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return post, err
	}

	postMap := helpers.FromJson[map[string]any](postJson)

	mediaCloudNames := postMap["media_cloud_names"].([]any)

	var replacement []string

	for _, mcn := range mediaCloudNames {
		mcn := mcn.(string)

		var (
			blurPlchMcn string
			actualMcn   string
		)

		_, err = fmt.Sscanf(mcn, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
		if err != nil {
			helpers.LogError(err)
			return post, err
		}

		blurPlchUrl, err := getMediaurl(blurPlchMcn)
		if err != nil {
			return post, err
		}

		actualUrl, err := getMediaurl(actualMcn)
		if err != nil {
			return post, err
		}

		replacement = append(replacement, fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl))
	}

	postMap["media_urls"] = replacement

	delete(postMap, "media_cloud_names")

	return helpers.MapToStruct[T](postMap), nil
}

func GetComment[T any](ctx context.Context, commentId string) (comment T, err error) {
	commentJson, err := rdb().HGet(ctx, "comments", commentId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return comment, err
	}

	commentMap := helpers.FromJson[map[string]any](commentJson)

	attachmentCloudName := commentMap["attachment_cloud_name"].(string)

	var attachmentUrl string

	if attachmentCloudName != "" {
		attachmentUrl, err = getMediaurl(attachmentCloudName)
		if err != nil {
			return comment, err
		}
	}

	commentMap["attachment_url"] = attachmentUrl

	delete(commentMap, "attachment_cloud_name")

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
