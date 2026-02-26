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
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.FollowCursor = helpers.ParseInt(stmsg.Values["followCursor"].(string))

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
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				if len(notifications) > 0 {
					cache.StoreNewNotifications(pipe, ctx, notifications)

					cache.StoreUnreadNotifications(pipe, ctx, unreadNotifications)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {

				for followerUser, followingUser_score_Pairs := range userFollowings {
					cache.StoreUserFollowings(pipe, ctx, followerUser, followingUser_score_Pairs)
				}

				for followingUser, followerUser_score_Pairs := range userFollowers {
					cache.StoreUserFollowers(pipe, ctx, followingUser, followerUser_score_Pairs)
				}

				for user, notifId_score_Pairs := range userNotifications {
					cache.StoreUserNotifications(pipe, ctx, user, notifId_score_Pairs)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			for _, fn := range sendNotifEventMsgFuncs {
				go fn()
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
