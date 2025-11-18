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

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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

			// batch processing
			if len(notifications) > 0 {
				if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
					return
				}

				if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
					return
				}
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for followerUser, followingUser_stmsgId_Pairs := range userFollowings {
				eg.Go(func() error {
					followerUser, followingUser_stmsgId_Pairs := followerUser, followingUser_stmsgId_Pairs

					return cache.StoreUserFollowings(sharedCtx, followerUser, followingUser_stmsgId_Pairs)
				})
			}

			for followingUser, followerUser_stmsgId_Pairs := range userFollowers {
				eg.Go(func() error {
					followingUser, followerUser_stmsgId_Pairs := followingUser, followerUser_stmsgId_Pairs

					return cache.StoreUserFollowers(sharedCtx, followingUser, followerUser_stmsgId_Pairs)
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				eg.Go(func() error {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					return cache.StoreUserNotifications(sharedCtx, user, notifId_stmsgId_Pairs)
				})
			}

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
