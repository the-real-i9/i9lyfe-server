package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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
			var msgs []eventTypes.NewPostEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewPostEvent

				msg.OwnerUser = stmsg.Values["ownerUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostData = stmsg.Values["postData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.Hashtags = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["hashtags"].(string))
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			newPosts := []string{}

			userPosts := make(map[string][][2]string)

			userMentionedPosts := make(map[string][][2]string)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			newPostDBExtrasFuncs := []func() error{}

			fanOutPostFuncs := []func(){}

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				newPosts = append(newPosts, msg.PostId, msg.PostData)

				userPosts[msg.OwnerUser] = append(userPosts[msg.OwnerUser], [2]string{msg.PostId, stmsgIds[i]})

				newPostDBExtrasFuncs = append(newPostDBExtrasFuncs, func() error {
					return postModel.NewPostExtras(ctx, msg.PostId, msg.Mentions, msg.Hashtags)
				})

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

					notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
					unreadNotifications = append(unreadNotifications, notifUniqueId)

					userNotifications[mu] = append(userNotifications[mu], [2]string{notifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						uimu, err := cache.GetUser[UITypes.ClientUser](context.Background(), msg.OwnerUser)
						if err != nil {
							return
						}

						uimu.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uimu.ProfilePicUrl)

						notif["unread"] = true
						notif["details"].(map[string]any)["mentioning_user"] = uimu

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notif,
						})
					})
				}

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.PostId)
				})
			}

			// batch processing
			if err := cache.StoreNewPosts(ctx, newPosts); err != nil {
				return
			}

			if len(notifications) > 0 {
				if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
					return
				}

				if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
					return
				}
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for user, postId_stmsgId_Pairs := range userPosts {
				eg.Go(func() error {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					return cache.StoreUserPosts(sharedCtx, user, postId_stmsgId_Pairs)
				})
			}

			for user, postId_stmsgId_Pairs := range userMentionedPosts {
				eg.Go(func() error {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					return cache.StoreUserMentionedPosts(sharedCtx, user, postId_stmsgId_Pairs)
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				eg.Go(func() error {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					return cache.StoreUserNotifications(sharedCtx, user, notifId_stmsgId_Pairs)
				})
			}

			for _, fn := range newPostDBExtrasFuncs {
				eg.Go(func() error {
					fn := fn

					return fn()
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
