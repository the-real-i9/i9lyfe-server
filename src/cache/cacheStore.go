package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func StoreUserFeedPosts(pipe redis.Pipeliner, ctx context.Context, user string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAddGT(ctx, fmt.Sprintf("user:%s:feed", user), members...)
}
