package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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
				msg.CHECursor = helpers.FromJson[int64](stmsg.Values["cheCursor"].(string))

				msgs = append(msgs, msg)
			}

			newMessageEntries := []string{}

			newUserChats := make(map[string][]string)

			userChats := make(map[string]map[string]float64)

			updatedFromUserChats := make(map[string]map[string]float64)

			chatMessages := make(map[string][][2]any)

			// batch data for batch processing
			for _, msg := range msgs {
				newMessageEntries = append(newMessageEntries, msg.CHEId, msg.MsgData)

				if msg.FirstFromUser {
					newUserChats[msg.FromUser] = append(newUserChats[msg.FromUser], msg.ToUser, helpers.ToJson(map[string]any{"partner_user": msg.ToUser}))

					if userChats[msg.FromUser] == nil {
						userChats[msg.FromUser] = make(map[string]float64)
					}

					userChats[msg.FromUser][msg.ToUser] = float64(msg.CHECursor)
				} else {

					if updatedFromUserChats[msg.FromUser] == nil {
						updatedFromUserChats[msg.FromUser] = make(map[string]float64)
					}

					updatedFromUserChats[msg.FromUser][msg.ToUser] = float64(msg.CHECursor)
				}

				if msg.FirstToUser {
					newUserChats[msg.ToUser] = append(newUserChats[msg.ToUser], msg.FromUser, helpers.ToJson(map[string]any{"partner_user": msg.FromUser}))

					if userChats[msg.ToUser] == nil {
						userChats[msg.ToUser] = make(map[string]float64)
					}

					userChats[msg.ToUser][msg.FromUser] = float64(msg.CHECursor)
				}

				chatMessages[msg.FromUser+" "+msg.ToUser] = append(chatMessages[msg.FromUser+" "+msg.ToUser], [2]any{msg.CHEId, float64(msg.CHECursor)})
			}

			// batch processing
			if err := cache.StoreChatHistoryEntries(ctx, newMessageEntries); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUser, partnerUserWithChatInfoPairs := range newUserChats {
				eg.Go(func() error {
					ownerUser, partnerUserWithChatInfoPairs := ownerUser, partnerUserWithChatInfoPairs

					return cache.StoreUserChats(sharedCtx, ownerUser, partnerUserWithChatInfoPairs)
				})
			}

			for ownerUser, partnerUser_score_Pairs := range userChats {
				eg.Go(func() error {
					ownerUser, partnerUser_score_Pairs := ownerUser, partnerUser_score_Pairs

					return cache.StoreUserChatsSorted(sharedCtx, ownerUser, partnerUser_score_Pairs)
				})
			}

			for ownerUser, partnerUser_score_Pairs := range updatedFromUserChats {
				eg.Go(func() error {
					ownerUser, partnerUser_score_Pairs := ownerUser, partnerUser_score_Pairs

					return cache.StoreUserChatsSorted(sharedCtx, ownerUser, partnerUser_score_Pairs)
				})
			}

			for ownerUserPartnerUser, CHEId_score_Pairs := range chatMessages {
				eg.Go(func() error {
					ownerUserPartnerUser, CHEId_score_Pairs := ownerUserPartnerUser, CHEId_score_Pairs

					var ownerUser, partnerUser string

					fmt.Sscanf(ownerUserPartnerUser, "%s %s", &ownerUser, &partnerUser)

					return cache.StoreUserChatHistory(sharedCtx, ownerUser, partnerUser, CHEId_score_Pairs)
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
