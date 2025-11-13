package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func msgReactionsRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "msg_reactions_removed"
		groupName    = "msg_reaction_removed_listeners"
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
			var msgs []eventTypes.MsgReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.MsgReactionRemovedEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.ToMsgId = stmsg.Values["toMsgId"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)

				msgs = append(msgs, msg)

			}

			msgReactionEntriesRemoved := []string{}

			chatMsgReactionsRemoved := make(map[[2]string][][2]string)

			msgReactionsRemoved := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				msgReactionEntriesRemoved = append(msgReactionEntriesRemoved, msg.CHEId)

				chatMsgReactionsRemoved[[2]string{msg.FromUser, msg.ToUser}] = append(chatMsgReactionsRemoved[[2]string{msg.FromUser, msg.ToUser}], [2]string{msg.CHEId, stmsgIds[i]})

				msgReactionsRemoved[msg.ToMsgId] = append(msgReactionsRemoved[msg.ToMsgId], [2]string{msg.FromUser, stmsgIds[i]})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cache.RemoveChatHistoryEntries(ctx, msgReactionEntriesRemoved); err != nil {
				return
			}

			for ownerUserPartnerUser, CHEId_stmsgId_Pairs := range chatMsgReactionsRemoved {
				wg.Go(func() {
					ownerUserPartnerUser, CHEId_stmsgId_Pairs := ownerUserPartnerUser, CHEId_stmsgId_Pairs

					CHEIds := []any{}

					for _, pair := range CHEId_stmsgId_Pairs {
						CHEIds = append(CHEIds, pair[0])
					}

					if err := cache.RemoveUserChatHistory(ctx, ownerUserPartnerUser, CHEIds); err != nil {
						for _, pair := range CHEId_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for msgId, user_stmsgId_Pairs := range msgReactionsRemoved {
				wg.Go(func() {
					msgId, user_stmsgId_Pairs := msgId, user_stmsgId_Pairs

					reactorUsers := []string{}

					for _, user_stmsgId_Pair := range user_stmsgId_Pairs {
						reactorUsers = append(reactorUsers, user_stmsgId_Pair[0])
					}

					if err := cache.RemoveMsgReactions(ctx, msgId, reactorUsers); err != nil {
						for _, pair := range user_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
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
