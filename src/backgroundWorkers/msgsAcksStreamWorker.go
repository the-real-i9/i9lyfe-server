package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func msgAcksStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "msgs_acks"
		groupName    = "msgs_ack_listeners"
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
			var msgs []eventTypes.MsgsAckEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.MsgsAckEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.CHEIdList = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["cheIdList"].(string))
				msg.Ack = stmsg.Values["ack"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))
				msg.ChatCursor = helpers.FromJson[int64](stmsg.Values["chatCursor"].(string))

				msgs = append(msgs, msg)

			}

			ackMessages := [][3]any{}

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedFromUserChats := make(map[string]map[string]float64)

			// batch data for batch processing
			for _, msg := range msgs {

				for _, cheId := range msg.CHEIdList {
					ackMessages = append(ackMessages, [3]any{cheId, msg.Ack, msg.At})
				}

				if msg.Ack == "delivered" {
					if updatedFromUserChats[msg.FromUser] == nil {
						updatedFromUserChats[msg.FromUser] = make(map[string]float64)
					}

					updatedFromUserChats[msg.FromUser][msg.ToUser] = float64(msg.ChatCursor)

					if userChatUnreadMsgs[msg.FromUser] == nil {
						userChatUnreadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, cheId := range msg.CHEIdList {
						userChatUnreadMsgs[msg.FromUser][msg.ToUser] = append(userChatUnreadMsgs[msg.FromUser][msg.ToUser], cheId)
					}
				}

				if msg.Ack == "read" {
					if userChatReadMsgs[msg.FromUser] == nil {
						userChatReadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, cheId := range msg.CHEIdList {
						userChatReadMsgs[msg.FromUser][msg.ToUser] = append(userChatReadMsgs[msg.FromUser][msg.ToUser], cheId)
					}
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for _, msgId_ack_ackAt := range ackMessages {

				eg.Go(func() error {
					msgId, ack, ackAt := msgId_ack_ackAt[0], msgId_ack_ackAt[1], msgId_ack_ackAt[2]

					return cache.UpdateMessage(sharedCtx, msgId.(string), map[string]any{
						"delivery_status":         ack,
						fmt.Sprintf("%s_at", ack): ackAt.(int64),
					})
				})
			}

			for ownerUser, partnerUser_score_Pairs := range updatedFromUserChats {
				eg.Go(func() error {
					ownerUser, partnerUser_score_Pairs := ownerUser, partnerUser_score_Pairs

					return cache.StoreUserChatsSorted(sharedCtx, ownerUser, partnerUser_score_Pairs)
				})
			}

			for ownerUser, partnerUser_unreadMsgs_Map := range userChatUnreadMsgs {
				eg.Go(func() error {
					ownerUser, partnerUser_unreadMsgs_Map := ownerUser, partnerUser_unreadMsgs_Map

					for partnerUser, unreadMsgs := range partnerUser_unreadMsgs_Map {
						if err := cache.StoreUserChatUnreadMsgs(sharedCtx, ownerUser, partnerUser, unreadMsgs); err != nil {
							return err
						}
					}

					return nil
				})
			}

			for ownerUser, partnerUser_readMsgs_Map := range userChatReadMsgs {
				eg.Go(func() error {
					ownerUser, partnerUser_readMsgs_Map := ownerUser, partnerUser_readMsgs_Map

					for partnerUser, readMsgs := range partnerUser_readMsgs_Map {
						if err := cache.RemoveUserChatUnreadMsgs(sharedCtx, ownerUser, partnerUser, readMsgs); err != nil {
							return err
						}
					}

					return nil
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
