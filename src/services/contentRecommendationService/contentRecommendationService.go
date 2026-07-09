package contentRecommendationService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/services/sseService"
	"i9lyfe/src/types/appTypes"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func FanOutPostToFollowers(ctx context.Context, postId string, postCursor float64, user string) error {

	var nextCursor int64

	for {
		type f struct {
			Username string `db:"follower_username"`
			Cursor   int64  `db:"cursor_"`
		}
		// pull user followers from DB (cursor based)
		followers, err := pgDB.QueryRowsType[f](
			ctx,
			/* sql */ `
			SELECT follower_username, cursor_ FROM follows
			WHERE following_username = $1 AND cursor_ > $2
			`, user, nextCursor,
		)
		if err != nil {
			helpers.LogError(err)
			return err
		}

		if len(followers) == 0 {
			break
		}

		_, err = rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, f := range followers {
				cache.StoreUserFeedPosts(pipe, ctx, f.Username, [][2]any{{postId, postCursor}})
			}

			return nil
		})
		if err != nil {
			helpers.LogError(err)
			return err
		}

		for _, f := range followers {
			sseService.SendEventMsg(f.Username, appTypes.ServerEventMsg{
				Event: "new feed posts",
				Data:  nil,
			})
		}

		nextCursor = (followers[len(followers)-1]).Cursor
	}

	return nil
}
