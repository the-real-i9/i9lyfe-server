package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func usersFollowedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "users_followed"
		groupName    = "user_followed_listeners"
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
			var msgs []eventTypes.UserFollowEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.UserFollowEvent

				msg.FollowerUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["followerUser"].(string))
				msg.FollowingUser = stmsg.Values["followingUser"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)
			}

			userFollowers := make(map[string][][2]string)
			userFollowings := make(map[string][][2]string)

			notifications := []string{}

			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {

				userFollowings[msg.FollowerUser.Username] = append(userFollowings[msg.FollowerUser.Username], [2]string{msg.FollowingUser, stmsgIds[i]})

				userFollowers[msg.FollowingUser] = append(userFollowers[msg.FollowingUser], [2]string{msg.FollowerUser.Username, stmsgIds[i]})

				notifUniqueId := fmt.Sprintf("user_%s_follows_user_%s", msg.FollowerUser.Username, msg.FollowingUser)
				notif := helpers.BuildNotification(notifUniqueId, "user_follow", msg.At, map[string]any{
					"follower_user": msg.FollowerUser.Username,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.FollowingUser] = append(userNotifications[msg.FollowingUser], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["unread"] = true
					notif["details"].(map[string]any)["follower_user"] = msg.FollowerUser

					realtimeService.SendEventMsg(msg.FollowingUser, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notif,
					})
				})

			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing
			if len(notifications) > 0 {
				if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
					return
				}

				if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
					return
				}
			}

			for followerUser, followingUser_stmsgId_Pairs := range userFollowings {
				wg.Go(func() {
					followerUser, followingUser_stmsgId_Pairs := followerUser, followingUser_stmsgId_Pairs

					if err := cache.StoreUserFollowings(ctx, followerUser, followingUser_stmsgId_Pairs); err != nil {
						for _, pair := range followingUser_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for followingUser, followerUser_stmsgId_Pairs := range userFollowers {
				wg.Go(func() {
					followingUser, followerUser_stmsgId_Pairs := followingUser, followerUser_stmsgId_Pairs

					if err := cache.StoreUserFollowers(ctx, followingUser, followerUser_stmsgId_Pairs); err != nil {
						for _, pair := range followerUser_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					if err := cache.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs); err != nil {
						for _, pair := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

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
