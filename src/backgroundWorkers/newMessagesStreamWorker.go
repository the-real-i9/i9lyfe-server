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

func newMessagesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_messages"
		groupName    = "new_message_listeners"
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
			var msgs []eventTypes.NewMessageEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewMessageEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.MsgData = stmsg.Values["msgData"].(string)

				msgs = append(msgs, msg)
			}

			newMessageEntries := []string{}

			chatMessages := make(map[[2]string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				newMessageEntries = append(newMessageEntries, msg.CHEId, msg.MsgData)

				chatMessages[[2]string{msg.FromUser, msg.ToUser}] = append(chatMessages[[2]string{msg.FromUser, msg.ToUser}], [2]string{msg.CHEId, stmsgIds[i]})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cache.StoreChatHistoryEntries(ctx, newMessageEntries); err != nil {
				return
			}

			for ownerUserPartnerUser, CHEId_stmsgId_Pairs := range chatMessages {
				wg.Go(func() {
					ownerUserPartnerUser, CHEId_stmsgId_Pairs := ownerUserPartnerUser, CHEId_stmsgId_Pairs

					if err := cache.StoreUserChatHistory(ctx, ownerUserPartnerUser, CHEId_stmsgId_Pairs); err != nil {
						for _, d := range CHEId_stmsgId_Pairs {
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
