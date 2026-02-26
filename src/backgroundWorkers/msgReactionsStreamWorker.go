package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

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
				msg.CHECursor = helpers.ParseInt(stmsg.Values["cheCursor"].(string))

				msgs = append(msgs, msg)

			}

			newMsgReactionEntries := []string{}

			chatMsgReactions := make(map[string][][2]any)

			msgReactions := make(map[string][]string)

			// batch data for batch processing
			for _, msg := range msgs {
				newMsgReactionEntries = append(newMsgReactionEntries, msg.CHEId, msg.RxnData)

				chatMsgReactions[msg.FromUser+" "+msg.ToUser] = append(chatMsgReactions[msg.FromUser+" "+msg.ToUser], [2]any{msg.CHEId, float64(msg.CHECursor)})

				msgReactions[msg.ToMsgId] = append(msgReactions[msg.ToMsgId], msg.FromUser, msg.Emoji)
			}

			// batch processing
			if err := cache.StoreChatHistoryEntries(ctx, newMsgReactionEntries); err != nil {
				return
			}

			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for ownerUserPartnerUser, CHEId_score_Pairs := range chatMsgReactions {

					var ownerUser, partnerUser string

					fmt.Sscanf(ownerUserPartnerUser, "%s %s", &ownerUser, &partnerUser)

					cache.StoreUserChatHistory(pipe, ctx, ownerUser, partnerUser, CHEId_score_Pairs)
				}

				for msgId, userWithEmojiPairs := range msgReactions {
					cache.StoreMsgReactions(pipe, ctx, msgId, userWithEmojiPairs)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
