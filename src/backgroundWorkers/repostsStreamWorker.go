package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cacheService"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func repostsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "reposts"
		groupName    = "repost_listeners"
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

			var msgs []eventTypes.RepostEvent
			helpers.ToStruct(stmsgValues, &msgs)

			msgsLen := len(msgs)

			reposts := make(map[string]any, msgsLen)

			userPosts := make(map[string][][2]string)

			notifications := make(map[any]any)

			userNotifications := make(map[string][][2]string)

			fanOutPostFuncs := make([]func(), msgsLen)

			sendNotifEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for i, msg := range msgs {
				reposts[msg.RepostId] = msg.RepostData

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.RepostId)
				})

				if msg.ReposterUser == msg.PostOwner {
					continue
				}

				userPosts[msg.ReposterUser] = append(userPosts[msg.ReposterUser], [2]string{msg.RepostId, stmsgIds[i]})

				notifUniqueId := fmt.Sprintf("user_%s_reposted_post_%s", msg.ReposterUser, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "repost", msg.At, map[string]any{
					"reposted_post_id": msg.PostId,
					"repost_id":        msg.RepostId,
					"reposter_user":    msg.ReposterUser,
				})

				notifications[notifUniqueId] = helpers.ToJson(notif)

				userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["is_read"] = false

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  helpers.ToJson(notif),
					})
				})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cacheService.StoreNewPosts(ctx, reposts); err != nil {
				return
			}

			if err := cacheService.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			for user, postId_stmsgId_Pairs := range userPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs
					if err := cacheService.StoreUserPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					if err := cacheService.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs); err != nil {
						for _, d := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			go func() {
				for _, fn := range fanOutPostFuncs {
					fn()
				}
			}()

			go func() {
				for _, fn := range sendNotifEventMsgFuncs {
					fn()
				}
			}()

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
