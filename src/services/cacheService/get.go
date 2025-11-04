package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
)

// These functions builds the frontend struction of the
// requested data by querying relevant cache keys

func GetMPosts(postId ...string) []map[string]any {

	return nil
}

func GetPost(ctx context.Context, postId string) (map[string]any, error) {
	post, err := rdb.HGetAll(ctx, fmt.Sprintf("post:%s", postId)).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return helpers.Json2Map(helpers.ToJson(post)), nil
}
