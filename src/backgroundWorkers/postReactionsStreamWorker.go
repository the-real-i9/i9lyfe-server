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

func postReactionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_reactions"
		groupName    = "post_reaction_listeners"
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
			var msgs []eventTypes.PostReactionEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostReactionEvent

				msg.ReactorUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["reactorUser"].(string))
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			postReactions := make(map[string][][2]any)

			userReactedPosts := make(map[string][][2]string)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string, msgsLen)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				postReactions[msg.PostId] = append(postReactions[msg.PostId], [2]any{[]string{msg.ReactorUser.Username, msg.Emoji}, stmsgIds[i]})

				userReactedPosts[msg.ReactorUser.Username] = append(userReactedPosts[msg.ReactorUser.Username], [2]string{msg.PostId, stmsgIds[i]})

				if msg.PostOwner == msg.ReactorUser.Username {
					continue
				}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_post_%s", msg.ReactorUser.Username, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_post", msg.At, map[string]any{
					"to_post_id":   msg.PostId,
					"reactor_user": msg.ReactorUser.Username,
					"emoji":        msg.Emoji,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["unread"] = true
					notif["details"].(map[string]any)["reactor_user"] = msg.ReactorUser

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notif,
					})
				})
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
				return
			}

			for postId, userWithEmoji_stmsgId_Pairs := range postReactions {
				wg.Go(func() {
					postId, userWithEmoji_stmsgId_Pairs := postId, userWithEmoji_stmsgId_Pairs

					userWithEmojiPairs := [][]string{}

					for _, userWithEmoji_stmsgId_Pair := range userWithEmoji_stmsgId_Pairs {
						userWithEmojiPairs = append(userWithEmojiPairs, userWithEmoji_stmsgId_Pair[0].([]string))
					}

					if err := cache.StorePostReactions(ctx, postId, slices.Concat(userWithEmojiPairs...)); err != nil {
						for _, d := range userWithEmoji_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for postId := range postReactions {
					totalRxnsCount, err := cache.GetPostReactionsCount(ctx, postId)
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":                postId,
						"latest_reactions_count": totalRxnsCount,
					})
				}
			}()

			for user, postId_stmsgId_Pairs := range userReactedPosts {
				wg.Go(func() {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					if err := cache.StoreUserReactedPosts(ctx, user, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					err = cache.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs)
					if err != nil {
						for _, d := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
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
