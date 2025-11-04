package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cacheService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func postReactionRemovalStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_reaction_removal"
		groupName    = "post_reaction_removal_listeners"
		consumerName = "worker-1"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		// cache new post and publish to subscribers
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

			var msgs []eventTypes.RemovePostReactionEvent
			helpers.ToStruct(stmsgValues, &msgs)

			postReactionsRemoved := make(map[string][][2]string)

			postIdsUndoed := []any{}

			// batch data for batch processing
			for i, msg := range msgs {

				postReactionsRemoved[msg.ReactorUser] = append(postReactionsRemoved[msg.ReactorUser], [2]string{msg.PostId, stmsgIds[i]})

				postIdsUndoed = append(postIdsUndoed, msg.PostId)
			}

			// batch processing
			wg := new(sync.WaitGroup)
			failedStreamMsgIds := make(map[string]bool)

			for user, postId_stmsgId_Pairs := range postReactionsRemoved {
				postIds := []any{}
				stmsgIds := []string{}

				for _, postId_stmsgId_Pair := range postId_stmsgId_Pairs {
					postIds = append(postIds, postId_stmsgId_Pair[0])
					stmsgIds = append(stmsgIds, postId_stmsgId_Pair[1])
				}

				wg.Go(func() {
					if err := cacheService.RemoveReactionToPost(ctx, user, postIds); err != nil {
						for _, id := range stmsgIds {
							failedStreamMsgIds[id] = true
						}
					}
				})
			}

			for _, postId := range postIdsUndoed {
				go func() {
					latestCount, err := rdb.HLen(ctx, fmt.Sprintf("post:%s:reactions", postId)).Result()
					if err != nil {
						return
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":                postId,
						"latest_reactions_count": latestCount,
					})
				}()
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
