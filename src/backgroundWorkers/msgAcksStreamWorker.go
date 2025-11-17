package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func msgAcksStreamBgWorker(rdb *redis.Client) {
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
			var msgs []eventTypes.MsgAckEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.MsgAckEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.Ack = stmsg.Values["ack"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			ackMessages := [][4]any{}

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedFromUserChats := make(map[string]map[string]string)

			// batch data for batch processing
			for i, msg := range msgs {

				ackMessages = append(ackMessages, [4]any{msg.CHEId, msg.Ack, msg.At, stmsgIds[i]})

				if msg.Ack == "delivered" {
					updatedFromUserChats[msg.FromUser][msg.ToUser] = stmsgIds[i]

					userChatUnreadMsgs[msg.FromUser][msg.ToUser] = append(userChatUnreadMsgs[msg.FromUser][msg.ToUser], msg.CHEId)
				}

				if msg.Ack == "read" {
					userChatReadMsgs[msg.FromUser][msg.ToUser] = append(userChatReadMsgs[msg.FromUser][msg.ToUser], msg.CHEId)
				}
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			for _, CHEId_ack_ackAt_stmsgId := range ackMessages {

				wg.Go(func() {
					CHEId, ack, ackAt, stmsgId := CHEId_ack_ackAt_stmsgId[0], CHEId_ack_ackAt_stmsgId[1], CHEId_ack_ackAt_stmsgId[2], CHEId_ack_ackAt_stmsgId[3]

					if err := cache.UpdateMessage(ctx, CHEId.(string), map[string]any{
						"delivery_status":         ack,
						fmt.Sprintf("%s_at", ack): ackAt.(int64),
					}); err != nil {
						failedStreamMsgIds[stmsgId.(string)] = true
					}
				})
			}

			for ownerUser, partnerUser_stmsgId_Pairs := range updatedFromUserChats {
				wg.Go(func() {
					ownerUser, partnerUser_stmsgId_Pairs := ownerUser, partnerUser_stmsgId_Pairs

					if err := cache.StoreUserChatsSorted(ctx, ownerUser, partnerUser_stmsgId_Pairs); err != nil {
						for _, stmsgId := range partnerUser_stmsgId_Pairs {
							failedStreamMsgIds[stmsgId] = true
						}
					}
				})
			}

			for ownerUser, partnerUser_unreadMsgs_Map := range userChatUnreadMsgs {
				wg.Go(func() {
					ownerUser, partnerUser_unreadMsgs_Map := ownerUser, partnerUser_unreadMsgs_Map

					for partnerUser, unreadMsgs := range partnerUser_unreadMsgs_Map {
						if err := cache.StoreUserChatUnreadMsgs(ctx, ownerUser, partnerUser, unreadMsgs); err != nil {
							// signal error
						}
					}
				})
			}

			for ownerUser, partnerUser_readMsgs_Map := range userChatReadMsgs {
				wg.Go(func() {
					ownerUser, partnerUser_readMsgs_Map := ownerUser, partnerUser_readMsgs_Map

					for partnerUser, readMsgs := range partnerUser_readMsgs_Map {
						if err := cache.RemoveUserChatUnreadMsgs(ctx, ownerUser, partnerUser, readMsgs); err != nil {
							// signal error
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
