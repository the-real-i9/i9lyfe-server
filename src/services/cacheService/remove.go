package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
)

func RemoveReactionToPost(ctx context.Context, user string, postIds []any) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("reacted_post:%s:reactions", postIds...), user).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:reacted_posts", user), postIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
