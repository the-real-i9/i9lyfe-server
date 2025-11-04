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
		// cache new post and publish to subscribers
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

			newPosts := make(map[string][2]any, msgsLen)

			userPosts := make(map[string][][2]string)

			userMentionedPosts := make(map[string][][2]string)

			notifications := make(map[string][2]any)

			userNotifications := make(map[string][][2]string)

			newPostExtrasFuncs := make([][2]any, msgsLen)

			fanOutPostFuncs := make([]func(), msgsLen)

			sendEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for i, msg := range msgs {
				newPosts[msg.PostId] = [2]any{msg.PostData, stmsgIds[i]}

				userPosts[msg.OwnerUser] = append(userPosts[msg.OwnerUser], [2]string{msg.PostId, stmsgIds[i]})

				for _, mu := range msg.Mentions {
					userMentionedPosts[mu] = append(userMentionedPosts[mu], [2]string{msg.PostId, stmsgIds[i]})

					notifUniqueId := fmt.Sprintf("user_%s_mentioned_in_post_%s", mu, msg.PostId)
					notif := helpers.BuildNotification(notifUniqueId, "mention_in_post", msg.At, map[string]any{
						"in_post_id":      msg.PostId,
						"mentioning_user": msg.OwnerUser,
					})

					notifications[notifUniqueId] = [2]any{notif, stmsgIds[i]}

					userNotifications[mu] = append(userNotifications[mu], [2]string{notifUniqueId, stmsgIds[i]})

					sendEventMsgFuncs = append(sendEventMsgFuncs, func() {
						notif["notif"] = helpers.Json2Map(notif["notif"].(string))

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notif,
						})
					})
				}

				newPostExtrasFuncs = append(newPostExtrasFuncs, [2]any{func() error {
					return postModel.NewPostExtras(ctx, msg.PostId, msg.Mentions, msg.Hashtags)
				}, stmsgIds[i]})

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.PostId)
				})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			for postId, postData_stmsgId_Pair := range newPosts {
				wg.Go(func() {
					postId, postData, stmsgId := postId, postData_stmsgId_Pair[0], postData_stmsgId_Pair[1].(string)

					if err := cacheService.StoreNewPosts(ctx, postId, postData); err != nil {
						failedStreamMsgIds[stmsgId] = true
					}
				})
			}

			wg.Wait()

			for notifId, notif_stmsgId_Pair := range notifications {
				wg.Go(func() {
					notifId, notifData, stmsgId := notifId, notif_stmsgId_Pair[0], notif_stmsgId_Pair[1].(string)

					if err := cacheService.StoreNewNotifications(ctx, notifId, notifData); err != nil {
						failedStreamMsgIds[stmsgId] = true
					}
				})
			}

			wg.Wait()

			for k, postId_stmsgId_Pairs := range userPosts {
				wg.Go(func() {
					k, postId_stmsgId_Pairs := k, postId_stmsgId_Pairs
					if err := cacheService.StoreUserPosts(ctx, k, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for k, postId_stmsgId_Pairs := range userMentionedPosts {
				wg.Go(func() {
					k, postId_stmsgId_Pairs := k, postId_stmsgId_Pairs

					if err := cacheService.StoreUserMentionedPosts(ctx, k, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for k, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					k, notifId_stmsgId_Pairs := k, notifId_stmsgId_Pairs

					if err := cacheService.StoreUserNotifications(ctx, k, notifId_stmsgId_Pairs); err != nil {
						for _, d := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for _, fn_stmsgId_Pair := range newPostExtrasFuncs {
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
				for _, fn := range sendEventMsgFuncs {
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
