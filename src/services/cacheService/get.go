package cacheService

import (
	"context"
	"i9lyfe/src/helpers"
)

// These functions builds the frontend struction of the
// requested data by querying relevant cache keys

func GetMPosts(postId ...string) []map[string]any {

	return nil
}

func GetPost(ctx context.Context, postId string) (map[string]any, error) {
	post, err := rdb.HGet(ctx, "posts", postId).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return helpers.FromJson[map[string]any](post), nil
}
