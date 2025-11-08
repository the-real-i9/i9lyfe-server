package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cacheService"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func newPostsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_posts"
		groupName    = "new_post_listeners"
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

			var msgs []eventTypes.NewPostEvent
			helpers.ToStruct(stmsgValues, &msgs)

			msgsLen := len(msgs)

			newPosts := []string{}

			userPosts := make(map[string][][2]string)

			userMentionedPosts := make(map[string][][2]string)

			notifications := make(map[any]any)

			userNotifications := make(map[string][][2]string)

			newPostDBExtrasFuncs := make([][2]any, msgsLen)

			fanOutPostFuncs := make([]func(), msgsLen)

			sendNotifEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for i, msg := range msgs {
				newPosts = append(newPosts, msg.PostId, msg.PostData)

				userPosts[msg.OwnerUser] = append(userPosts[msg.OwnerUser], [2]string{msg.PostId, stmsgIds[i]})

				newPostDBExtrasFuncs = append(newPostDBExtrasFuncs, [2]any{func() error {
					return postModel.NewPostExtras(ctx, msg.PostId, msg.Mentions, msg.Hashtags)
				}, stmsgIds[i]})

				for _, mu := range msg.Mentions {
					userMentionedPosts[mu] = append(userMentionedPosts[mu], [2]string{msg.PostId, stmsgIds[i]})

					if mu == msg.OwnerUser {
						continue
					}

					notifUniqueId := fmt.Sprintf("user_%s_mentioned_in_post_%s", mu, msg.PostId)
					notif := helpers.BuildNotification(notifUniqueId, "mention_in_post", msg.At, map[string]any{
						"in_post_id":      msg.PostId,
						"mentioning_user": msg.OwnerUser,
					})

					notifications[notifUniqueId] = helpers.ToJson(notif)

					userNotifications[mu] = append(userNotifications[mu], [2]string{notifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						notif["is_read"] = false

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  helpers.ToJson(notif),
						})
					})
				}

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.PostId)
				})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cacheService.StoreNewPosts(ctx, newPosts); err != nil {
				return
			}

			if err := cacheService.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			for user, postId_stmsgId_Pairs := range userPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs
					if err := cacheService.StoreUserPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, pair := range postId_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for user, postId_stmsgId_Pairs := range userMentionedPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					if err := cacheService.StoreUserMentionedPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, pair := range postId_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					if err := cacheService.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs); err != nil {
						for _, pair := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for _, fn_stmsgId_Pair := range newPostDBExtrasFuncs {
				wg.Go(func() {
					fn, stmsgId := fn_stmsgId_Pair[0].(func() error), fn_stmsgId_Pair[1].(string)

					if err := fn(); err != nil {
						failedStreamMsgIds[stmsgId] = true
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
