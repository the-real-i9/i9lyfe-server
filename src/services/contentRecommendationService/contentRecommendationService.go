package contentRecommendationService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"
	"i9lyfe/src/services/realtimeService"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func FanOutPostToFollowers(postId string, postCursor float64, user string) {
	ctx := context.Background()

	var nextCursor uint64

	for {
		followers, cursor, err := rdb().ZScan(ctx, fmt.Sprintf("user:%s:followers", user), nextCursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			break
		}

		go func(followers []string) {
			ctx := context.Background()

			_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for _, fuser := range followers {
					cache.StoreUserFeedPosts(pipe, ctx, fuser, [][2]any{{postId, postCursor}})
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			for _, fuser := range followers {
				postUI, err := modelHelpers.BuildPostUIFromCache(ctx, postId, fuser)
				if err != nil {
					helpers.LogError(err)
					continue
				}

				realtimeService.SendNewFeedPostEventMsg(fuser, postUI)
			}
		}(followers)

		if cursor == 0 {
			break
		}

		nextCursor = cursor
	}
}
