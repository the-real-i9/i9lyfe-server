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

func postUnsavesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_unsaves"
		groupName    = "post_unsave_listeners"
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
			var msgs []eventTypes.PostUnsaveEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostUnsaveEvent

				msg.SaverUser = stmsg.Values["saverUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)

				msgs = append(msgs, msg)

			}

			postUnsaves := make(map[string][]any)

			userUnsavedPosts := make(map[string][]any)

			// batch data for batch processing
			for _, msg := range msgs {

				postUnsaves[msg.PostId] = append(postUnsaves[msg.PostId], msg.SaverUser)

				userUnsavedPosts[msg.SaverUser] = append(userUnsavedPosts[msg.SaverUser], msg.PostId)
			}

			// batch processing
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for postId, users := range postUnsaves {
					cache.RemovePostSaves(pipe, ctx, postId, users)
				}

				for user, postIds := range userUnsavedPosts {
					cache.RemoveUserSavedPosts(pipe, ctx, user, postIds)
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
					for postId := range postUnsaves {
						postId_IntCmd[postId] = cache.GetPostSavesCount(pipe, ctx, postId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for postId, lc := range postId_IntCmd {
					latestCount, err := lc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":            postId,
						"latest_saves_count": latestCount,
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
