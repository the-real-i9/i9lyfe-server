package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
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
			var stmsgValues []map[string]any

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)
				stmsgValues = append(stmsgValues, stmsg.Values)

			}

			var msgs []eventTypes.PostSaveEvent
			helpers.ToStruct(stmsgValues, &msgs)

			postSaves := make(map[string][][2]any)

			userSavedPosts := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				postSaves[msg.PostId] = append(postSaves[msg.PostId], [2]any{msg.SaverUser, stmsgIds[i]})

				userSavedPosts[msg.SaverUser] = append(userSavedPosts[msg.SaverUser], [2]string{msg.PostId, stmsgIds[i]})
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			for postId, user_stmsgId_Pairs := range postSaves {
				wg.Go(func() {
					postId, user_stmsgId_Pairs := postId, user_stmsgId_Pairs

					saverUsers := []any{}

					for _, user_stmsgId_Pair := range user_stmsgId_Pairs {
						saverUsers = append(saverUsers, user_stmsgId_Pair[0].(string))
					}

					if err := cache.StorePostSaves(ctx, postId, saverUsers); err != nil {
						for _, d := range user_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for postId := range postSaves {
					totalRxnsCount, err := cache.GetPostSavesCount(ctx, postId)
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":            postId,
						"latest_saves_count": totalRxnsCount,
					})
				}
			}()

			for user, postId_stmsgId_Pairs := range userSavedPosts {

				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					if err := cache.StoreUserSavedPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			wg.Wait()

			stmsgIds = slices.DeleteFunc(stmsgIds, func(stmsgId string) bool {
				return failedStreamMsgIds[stmsgId]
			})

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
