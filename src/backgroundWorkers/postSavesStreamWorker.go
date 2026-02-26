package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

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
			var msgs []eventTypes.PostSaveEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostSaveEvent

				msg.SaverUser = stmsg.Values["saverUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.SaveCursor = helpers.ParseInt(stmsg.Values["saveCursor"].(string))

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
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for postId, saverUsers := range postSaves {
					cache.StorePostSaves(pipe, ctx, postId, saverUsers)
				}

				for user, postId_score_Pairs := range userSavedPosts {
					cache.StoreUserSavedPosts(pipe, ctx, user, postId_score_Pairs)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			go func() {
				ctx := context.Background()
				postId_IntCmd := make(map[string]*redis.IntCmd)

				_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					for postId := range postSaves {
						postId_IntCmd[postId] = cache.GetPostSavesCount(pipe, ctx, postId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for postId, lc := range postId_IntCmd {
					totalSavesCount, err := lc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":            postId,
						"latest_saves_count": totalSavesCount,
					})
				}
			}()

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
