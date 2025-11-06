package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
)

func RemovePostReactions(ctx context.Context, postId string, users []string) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemovePostSaves(ctx context.Context, postId string, users []any) error {
	if err := rdb.SRem(ctx, fmt.Sprintf("saved_post:%s:saves", postId), users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveCommentReactions(ctx context.Context, commentId string, users []string) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserReactedPosts(ctx context.Context, user string, postIds []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:reacted_posts", user), postIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserSavedPosts(ctx context.Context, user string, postIds []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:saved_posts", user), postIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
