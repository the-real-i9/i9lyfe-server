package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func postSavesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_saves"
		groupName    = "post_save_listeners"
		consumerName = "worker-1"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		for {
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    500,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var msgs []eventTypes.PostSaveEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostSaveEvent

				msg.SaverUser = stmsg.Values["saverUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.SaveCursor = helpers.FromJson[int64](stmsg.Values["saveCursor"].(string))

				msgs = append(msgs, msg)

			}

			postSaves := make(map[string][]any)

			userSavedPosts := make(map[string][][2]any)

			// batch data for batch processing
			for _, msg := range msgs {
				postSaves[msg.PostId] = append(postSaves[msg.PostId], msg.SaverUser)

				userSavedPosts[msg.SaverUser] = append(userSavedPosts[msg.SaverUser], [2]any{msg.PostId, float64(msg.SaveCursor)})
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for postId, saverUsers := range postSaves {
				eg.Go(func() error {
					postId, saverUsers := postId, saverUsers

					if err := cache.StorePostSaves(sharedCtx, postId, saverUsers); err != nil {
						return err
					}

					go func() {
						ctx := context.Background()
						for postId := range postSaves {
							totalSavesCount, err := cache.GetPostSavesCount(ctx, postId)
							if err != nil {
								continue
							}

							realtimeService.PublishPostMetric(ctx, map[string]any{
								"post_id":            postId,
								"latest_saves_count": totalSavesCount,
							})
						}
					}()

					return nil
				})
			}

			for user, postId_score_Pairs := range userSavedPosts {
				eg.Go(func() error {
					user, postId_score_Pairs := user, postId_score_Pairs

					return cache.StoreUserSavedPosts(sharedCtx, user, postId_score_Pairs)
				})
			}

			if eg.Wait() != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
