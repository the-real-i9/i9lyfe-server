package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/commentModel"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func commentCommentsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "comment_comments"
		groupName    = "comment_comment_listeners"
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
			var msgs []eventTypes.CommentCommentEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.CommentCommentEvent

				msg.CommenterUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["commenterUser"].(string))
				msg.ParentCommentId = stmsg.Values["parentCommentId"].(string)
				msg.ParentCommentOwner = stmsg.Values["parentCommentOwner"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentData = stmsg.Values["commentData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)
			}

			newComments := []string{}

			commentComments := make(map[string][][2]string)

			newCommentDBExtrasFuncs := [][2]any{}

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				newComments = append(newComments, msg.CommentId, msg.CommentData)

				commentComments[msg.ParentCommentId] = append(commentComments[msg.ParentCommentId], [2]string{msg.CommentId, stmsgIds[i]})

				newCommentDBExtrasFuncs = append(newCommentDBExtrasFuncs, [2]any{func() error {
					return commentModel.CommentOnExtras(ctx, msg.CommentId, msg.Mentions)
				}, stmsgIds[i]})

				if msg.ParentCommentOwner == msg.CommenterUser.Username {

					cocNotifUniqueId := fmt.Sprintf("user_%s_comment_%s_on_comment_%s", msg.CommenterUser.Username, msg.CommentId, msg.ParentCommentId)
					cocNotif := helpers.BuildNotification(cocNotifUniqueId, "comment_on_comment", msg.At, map[string]any{
						"on_comment_id":  msg.CommentId,
						"commenter_user": msg.CommenterUser.Username,
						"comment_id":     msg.CommentId,
					})

					notifications = append(notifications, cocNotifUniqueId, helpers.ToJson(cocNotif))
					unreadNotifications = append(unreadNotifications, cocNotifUniqueId)

					userNotifications[msg.ParentCommentOwner] = append(userNotifications[msg.ParentCommentOwner], [2]string{cocNotifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						cocNotif["unread"] = true
						cocNotif["details"].(map[string]any)["commenter_user"] = msg.CommenterUser

						realtimeService.SendEventMsg(msg.ParentCommentOwner, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  cocNotif,
						})
					})
				}

				for _, mu := range msg.Mentions {
					if mu == msg.CommenterUser.Username {
						continue
					}

					micNotifUniqueId := fmt.Sprintf("user_%s_mentioned_in_comment_%s", mu, msg.CommentId)
					micNotif := helpers.BuildNotification(micNotifUniqueId, "mention_in_comment", msg.At, map[string]any{
						"in_comment_id":   msg.CommentId,
						"mentioning_user": msg.CommenterUser.Username,
					})

					notifications = append(notifications, micNotifUniqueId, helpers.ToJson(micNotif))
					unreadNotifications = append(unreadNotifications, micNotifUniqueId)

					userNotifications[mu] = append(userNotifications[mu], [2]string{micNotifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						micNotif["unread"] = true
						micNotif["details"].(map[string]any)["mentioning_user"] = msg.CommenterUser

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  micNotif,
						})
					})
				}
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			if err := cache.StoreNewComments(ctx, newComments); err != nil {
				return
			}

			if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
				return
			}

			for parentCommentId, commentId_stmsgId_Pairs := range commentComments {
				wg.Go(func() {
					parentCommentId, commentId_stmsgId_Pairs := parentCommentId, commentId_stmsgId_Pairs

					if err := cache.StoreCommentComments(ctx, parentCommentId, commentId_stmsgId_Pairs); err != nil {
						for _, d := range commentId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for parentCommentId := range commentComments {
					totalCommentsCount, err := cache.GetCommentCommentsCount(ctx, parentCommentId)
					if err != nil {
						continue
					}

					realtimeService.PublishCommentMetric(ctx, map[string]any{
						"comment_id":            parentCommentId,
						"latest_comments_count": totalCommentsCount,
					})
				}
			}()

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

			for _, fn_stmsgId_Pair := range newCommentDBExtrasFuncs {
				wg.Go(func() {
					fn, stmsgId := fn_stmsgId_Pair[0].(func() error), fn_stmsgId_Pair[1].(string)

					if err := fn(); err != nil {
						failedStreamMsgIds[stmsgId] = true
					}
				})
			}

			go func() {
				for _, fn := range sendNotifEventMsgFuncs {
					fn()
				}
			}()

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
