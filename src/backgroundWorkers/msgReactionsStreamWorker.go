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

func msgReactionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "msg_reactions"
		groupName    = "msg_reaction_listeners"
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
			var msgs []eventTypes.NewMsgReactionEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewMsgReactionEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.RxnData = stmsg.Values["rxnData"].(string)
				msg.ToMsgId = stmsg.Values["toMsgId"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)

				msgs = append(msgs, msg)

			}

			newMsgReactionEntries := []string{}

			chatMsgReactions := make(map[[2]string][][2]string)

			msgReactions := make(map[string][][2]any)

			// batch data for batch processing
			for i, msg := range msgs {
				newMsgReactionEntries = append(newMsgReactionEntries, msg.CHEId, msg.RxnData)

				chatMsgReactions[[2]string{msg.FromUser, msg.ToUser}] = append(chatMsgReactions[[2]string{msg.FromUser, msg.ToUser}], [2]string{msg.CHEId, stmsgIds[i]})

				msgReactions[msg.ToMsgId] = append(msgReactions[msg.ToMsgId], [2]any{[]string{msg.FromUser, msg.Emoji}, stmsgIds[i]})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cache.StoreChatHistoryEntries(ctx, newMsgReactionEntries); err != nil {
				return
			}

			for ownerUserPartnerUser, CHEId_stmsgId_Pairs := range chatMsgReactions {
				wg.Go(func() {
					ownerUserPartnerUser, CHEId_stmsgId_Pairs := ownerUserPartnerUser, CHEId_stmsgId_Pairs

					if err := cache.StoreUserChatHistory(ctx, ownerUserPartnerUser, CHEId_stmsgId_Pairs); err != nil {
						for _, d := range CHEId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for msgId, userWithEmoji_stmsgId_Pairs := range msgReactions {
				wg.Go(func() {
					msgId, userWithEmoji_stmsgId_Pairs := msgId, userWithEmoji_stmsgId_Pairs

					userWithEmojiPairs := [][]string{}

					for _, userWithEmoji_stmsgId_Pair := range userWithEmoji_stmsgId_Pairs {
						userWithEmojiPairs = append(userWithEmojiPairs, userWithEmoji_stmsgId_Pair[0].([]string))
					}

					if err := cache.StoreMsgReactions(ctx, msgId, slices.Concat(userWithEmojiPairs...)); err != nil {
						for _, d := range userWithEmoji_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
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
