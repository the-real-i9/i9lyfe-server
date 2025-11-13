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

func commentReactionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "comment_reactions"
		groupName    = "comment_reaction_listeners"
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
			var msgs []eventTypes.CommentReactionEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.CommentReactionEvent

				msg.ReactorUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["reactorUser"].(string))
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentOwner = stmsg.Values["commentOwner"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			commentReactions := make(map[string][][2]any)

			// having comment reactors separate, allows us to
			// paginate through the list of reactions on a comment
			commentReactors := make(map[string][][2]string)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string, msgsLen)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				commentReactions[msg.CommentId] = append(commentReactions[msg.CommentId], [2]any{[]string{msg.ReactorUser.Username, msg.Emoji}, stmsgIds[i]})

				commentReactors[msg.CommentId] = append(commentReactors[msg.CommentId], [2]string{msg.ReactorUser.Username, stmsgIds[i]})

				if msg.CommentOwner == msg.ReactorUser.Username {
					continue
				}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_comment_%s", msg.ReactorUser.Username, msg.CommentId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_comment", msg.At, map[string]any{
					"to_comment_id": msg.CommentId,
					"reactor_user":  msg.ReactorUser.Username,
					"emoji":         msg.Emoji,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.CommentOwner] = append(userNotifications[msg.CommentOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["unread"] = true
					notif["details"].(map[string]any)["reactor_user"] = msg.ReactorUser

					realtimeService.SendEventMsg(msg.CommentOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notif,
					})
				})
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			if len(notifications) > 0 {
				if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
					return
				}

				if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
					return
				}
			}

			for commentId, userWithEmoji_stmsgId_Pairs := range commentReactions {
				wg.Go(func() {
					commentId, userWithEmoji_stmsgId_Pairs := commentId, userWithEmoji_stmsgId_Pairs

					userWithEmojiPairs := [][]string{}

					for _, userWithEmoji_stmsgId_Pair := range userWithEmoji_stmsgId_Pairs {
						userWithEmojiPairs = append(userWithEmojiPairs, userWithEmoji_stmsgId_Pair[0].([]string))
					}

					if err := cache.StoreCommentReactions(ctx, commentId, slices.Concat(userWithEmojiPairs...)); err != nil {
						for _, d := range userWithEmoji_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for commentId := range commentReactions {
					totalRxnsCount, err := cache.GetCommentReactionsCount(ctx, commentId)
					if err != nil {
						continue
					}

					realtimeService.PublishCommentMetric(ctx, map[string]any{
						"comment_id":             commentId,
						"latest_reactions_count": totalRxnsCount,
					})
				}
			}()

			for commentId, rUser_stmsgId_Pairs := range commentReactors {
				wg.Go(func() {
					commentId, rUser_stmsgId_Pairs := commentId, rUser_stmsgId_Pairs

					if err := cache.StoreCommentReactors(ctx, commentId, rUser_stmsgId_Pairs); err != nil {
						for _, d := range rUser_stmsgId_Pairs {
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
