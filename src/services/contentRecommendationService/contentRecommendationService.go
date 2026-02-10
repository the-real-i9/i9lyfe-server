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

func FanOutPostToFollowers(postId, user string) {
	ctx := context.Background()

	var nextCursor uint64

	for {
		followers, cursor, err := rdb().ZScan(ctx, fmt.Sprintf("user:%s:followers", user), nextCursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			break
		}

		for _, fuser := range followers {
			if err := cache.StoreUserFeedPost(ctx, fuser, postId); err != nil {
				helpers.LogError(err)
				continue
			}

			postUI, err := modelHelpers.BuildPostUIFromCache(ctx, postId, fuser)
			if err != nil {
				helpers.LogError(err)
				continue
			}

			postUI.Cursor, err = rdb().ZScore(ctx, fmt.Sprintf("user:%s:feed", fuser), postId).Result()
			if err != nil {
				helpers.LogError(err)
				continue
			}

			realtimeService.SendNewFeedPostEventMsg(fuser, postUI)
		}

		if cursor == 0 {
			break
		}

		nextCursor = cursor
	}
}
