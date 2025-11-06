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

func commentReactionRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "comment_reactions_removed"
		groupName    = "comment_reaction_removed_listeners"
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

			var msgs []eventTypes.CommentReactionRemovedEvent
			helpers.ToStruct(stmsgValues, &msgs)

			commentReactionsRemoved := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {

				commentReactionsRemoved[msg.CommentId] = append(commentReactionsRemoved[msg.CommentId], [2]string{msg.ReactorUser, stmsgIds[i]})

			}

			// batch processing
			wg := new(sync.WaitGroup)
			failedStreamMsgIds := make(map[string]bool)

			for commentId, user_stmsgId_Pairs := range commentReactionsRemoved {
				users := []string{}
				stmsgIds := []string{}

				for _, user_stmsgId_Pair := range user_stmsgId_Pairs {
					users = append(users, user_stmsgId_Pair[0])
					stmsgIds = append(stmsgIds, user_stmsgId_Pair[1])
				}

				wg.Go(func() {
					if err := cacheService.RemoveCommentReactions(ctx, commentId, users); err != nil {
						for _, id := range stmsgIds {
							failedStreamMsgIds[id] = true
						}
					}
				})
			}

			wg.Wait()

			for commentId := range commentReactionsRemoved {
				go func() {
					latestCount, err := rdb.HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId)).Result()
					if err != nil {
						return
					}

					realtimeService.PublishCommentMetric(ctx, map[string]any{
						"comment_id":             commentId,
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
