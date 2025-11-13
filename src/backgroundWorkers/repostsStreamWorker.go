package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
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
			var msgs []eventTypes.RepostEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.RepostEvent

				msg.ReposterUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["reposterUser"].(string))
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.RepostId = stmsg.Values["repostId"].(string)
				msg.RepostData = stmsg.Values["repostData"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			reposts := []string{}

			postReposts := make(map[string][][2]any)

			userRepostedPosts := make(map[string][][2]string)

			userPosts := make(map[string][][2]string)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			fanOutPostFuncs := []func(){}

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				reposts = append(reposts, msg.RepostId, msg.RepostData)

				postReposts[msg.PostId] = append(postReposts[msg.PostId], [2]any{msg.RepostId, stmsgIds[i]})

				userRepostedPosts[msg.ReposterUser.Username] = append(userRepostedPosts[msg.ReposterUser.Username], [2]string{msg.PostId, stmsgIds[i]})

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.RepostId)
				})

				if msg.ReposterUser.Username == msg.PostOwner {
					continue
				}

				userPosts[msg.ReposterUser.Username] = append(userPosts[msg.ReposterUser.Username], [2]string{msg.RepostId, stmsgIds[i]})

				notifUniqueId := fmt.Sprintf("user_%s_reposted_post_%s", msg.ReposterUser.Username, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "repost", msg.At, map[string]any{
					"reposted_post_id": msg.PostId,
					"reposter_user":    msg.ReposterUser.Username,
					"repost_id":        msg.RepostId,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["unread"] = true
					notif["details"].(map[string]any)["reposter_user"] = msg.ReposterUser

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notif,
					})
				})
			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if err := cache.StoreNewPosts(ctx, reposts); err != nil {
				return
			}

			if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
				return
			}

			for postId, repostId_stmsgId_Pairs := range postReposts {
				wg.Go(func() {
					postId, repostId_stmsgId_Pairs := postId, repostId_stmsgId_Pairs

					repostIds := []any{}

					for _, user_stmsgId_Pair := range repostId_stmsgId_Pairs {
						repostIds = append(repostIds, user_stmsgId_Pair[0].(string))
					}

					if err := cache.StorePostReposts(ctx, postId, repostIds); err != nil {
						for _, d := range repostId_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for postId := range postReposts {
					totalRepostsCount, err := cache.GetPostRepostsCount(ctx, postId)
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":              postId,
						"latest_reposts_count": totalRepostsCount,
					})
				}
			}()

			for user, postId_stmsgId_Pairs := range userPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs
					if err := cache.StoreUserPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for user, postId_stmsgId_Pairs := range userRepostedPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs
					if err := cache.StoreUserRepostedPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					if err := cache.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs); err != nil {
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
