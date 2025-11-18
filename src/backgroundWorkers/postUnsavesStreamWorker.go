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
			eg, sharedCtx := errgroup.WithContext(ctx)

			for postId, users := range postUnsaves {
				eg.Go(func() error {
					postId, users := postId, users

					return cache.RemovePostSaves(sharedCtx, postId, users)
				})
			}

			if eg.Wait() != nil {
				return
			}

			for postId := range postUnsaves {
				go func() {
					latestCount, err := cache.GetPostSavesCount(sharedCtx, postId)
					if err != nil {
						return
					}

					realtimeService.PublishPostMetric(sharedCtx, map[string]any{
						"post_id":            postId,
						"latest_saves_count": latestCount,
					})
				}()
			}

			for user, postIds := range userUnsavedPosts {
				eg.Go(func() error {
					user, postIds := user, postIds

					return cache.RemoveUserSavedPosts(sharedCtx, user, postIds)
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
