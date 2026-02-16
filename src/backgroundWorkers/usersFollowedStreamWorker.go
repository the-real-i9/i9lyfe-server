package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"
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

				msg.FollowerUser = stmsg.Values["followerUser"].(string)
				msg.FollowingUser = stmsg.Values["followingUser"].(string)
				msg.At = helpers.FromMsgPack[int64](stmsg.Values["at"].(string))
				msg.FollowCursor = helpers.FromMsgPack[int64](stmsg.Values["followCursor"].(string))

				msgs = append(msgs, msg)
			}

			userFollowers := make(map[string][][2]any)
			userFollowings := make(map[string][][2]any)

			notifications := []string{}

			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {

				userFollowings[msg.FollowerUser] = append(userFollowings[msg.FollowerUser], [2]any{msg.FollowingUser, float64(msg.FollowCursor)})

				userFollowers[msg.FollowingUser] = append(userFollowers[msg.FollowingUser], [2]any{msg.FollowerUser, float64(msg.FollowCursor)})

				notifUniqueId := fmt.Sprintf("user_%s_follows_user_%s", msg.FollowerUser, msg.FollowingUser)
				notif := helpers.BuildNotification(notifUniqueId, "user_follow", msg.At, map[string]any{
					"follower_user": msg.FollowerUser,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToMsgPack(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.FollowingUser] = append(userNotifications[msg.FollowingUser], [2]any{notifUniqueId, float64(msg.FollowCursor)})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), notifUniqueId)

					realtimeService.SendEventMsg(msg.FollowingUser, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notifSnippet,
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

			for followerUser, followingUser_score_Pairs := range userFollowings {
				eg.Go(func() error {
					followerUser, followingUser_score_Pairs := followerUser, followingUser_score_Pairs

					return cache.StoreUserFollowings(sharedCtx, followerUser, followingUser_score_Pairs)
				})
			}

			for followingUser, followerUser_score_Pairs := range userFollowers {
				eg.Go(func() error {
					followingUser, followerUser_score_Pairs := followingUser, followerUser_score_Pairs

					return cache.StoreUserFollowers(sharedCtx, followingUser, followerUser_score_Pairs)
				})
			}

			for user, notifId_score_Pairs := range userNotifications {
				eg.Go(func() error {
					user, notifId_score_Pairs := user, notifId_score_Pairs

					return cache.StoreUserNotifications(sharedCtx, user, notifId_score_Pairs)
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
