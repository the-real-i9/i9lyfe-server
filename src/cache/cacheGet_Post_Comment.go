package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

func GetPost[T any](ctx context.Context, postId string) (post T, err error) {
	postMsgPack, err := rdb().HGet(ctx, "posts", postId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return post, err
	}

	return helpers.FromMsgPack[T](postMsgPack), nil
}

func GetComment[T any](ctx context.Context, commentId string) (comment T, err error) {
	commentMsgPack, err := rdb().HGet(ctx, "comments", commentId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return comment, err
	}

	return helpers.FromMsgPack[T](commentMsgPack), nil
}

func GetPostReactionsCount(pipe redis.Pipeliner, ctx context.Context, postId string) *redis.IntCmd {
	return pipe.HLen(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId))
}

func GetPostCommentsCount(pipe redis.Pipeliner, ctx context.Context, postId string) *redis.IntCmd {
	return pipe.ZCard(ctx, fmt.Sprintf("commented_post:%s:comments", postId))
}

func GetPostSavesCount(pipe redis.Pipeliner, ctx context.Context, postId string) *redis.IntCmd {
	return pipe.SCard(ctx, fmt.Sprintf("saved_post:%s:saves", postId))
}

func GetPostRepostsCount(pipe redis.Pipeliner, ctx context.Context, postId string) *redis.IntCmd {
	return pipe.SCard(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId))
}

func GetCommentReactionsCount(pipe redis.Pipeliner, ctx context.Context, commentId string) *redis.IntCmd {
	return pipe.HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId))
}

func GetCommentCommentsCount(pipe redis.Pipeliner, ctx context.Context, parentCommentId string) *redis.IntCmd {
	return pipe.ZCard(ctx, fmt.Sprintf("commented_comment:%s:comments", parentCommentId))
}
