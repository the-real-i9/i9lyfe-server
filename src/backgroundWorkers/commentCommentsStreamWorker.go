package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/commentModel"
	"i9lyfe/src/models/modelHelpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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

				msg.CommenterUser = stmsg.Values["commenterUser"].(string)
				msg.ParentCommentId = stmsg.Values["parentCommentId"].(string)
				msg.ParentCommentOwner = stmsg.Values["parentCommentOwner"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentData = stmsg.Values["commentData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))
				msg.Score = helpers.FromJson[float64](stmsg.Values["score"].(string))

				msgs = append(msgs, msg)
			}

			newComments := []string{}

			commentComments := make(map[string][][2]any)

			newCommentDBExtrasFuncs := []func() error{}

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				newComments = append(newComments, msg.CommentId, msg.CommentData)

				commentComments[msg.ParentCommentId] = append(commentComments[msg.ParentCommentId], [2]any{msg.CommentId, msg.Score})

				newCommentDBExtrasFuncs = append(newCommentDBExtrasFuncs, func() error {
					return commentModel.CommentOnExtras(ctx, msg.CommentId, msg.Mentions)
				})

				if msg.ParentCommentOwner != msg.CommenterUser {

					cocNotifUniqueId := fmt.Sprintf("user_%s_comment_%s_on_comment_%s", msg.CommenterUser, msg.CommentId, msg.ParentCommentId)
					cocNotif := helpers.BuildNotification(cocNotifUniqueId, "comment_on_comment", msg.At, map[string]any{
						"on_comment_id":  msg.CommentId,
						"commenter_user": msg.CommenterUser,
						"comment_id":     msg.CommentId,
					})

					notifications = append(notifications, cocNotifUniqueId, helpers.ToJson(cocNotif))
					unreadNotifications = append(unreadNotifications, cocNotifUniqueId)

					userNotifications[msg.ParentCommentOwner] = append(userNotifications[msg.ParentCommentOwner], [2]string{cocNotifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), cocNotifUniqueId)

						realtimeService.SendEventMsg(msg.ParentCommentOwner, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notifSnippet,
						})
					})
				}

				for _, mu := range msg.Mentions {
					if mu == msg.CommenterUser {
						continue
					}

					micNotifUniqueId := fmt.Sprintf("user_%s_mentioned_in_comment_%s", mu, msg.CommentId)
					micNotif := helpers.BuildNotification(micNotifUniqueId, "mention_in_comment", msg.At, map[string]any{
						"in_comment_id":   msg.CommentId,
						"mentioning_user": msg.CommenterUser,
					})

					notifications = append(notifications, micNotifUniqueId, helpers.ToJson(micNotif))
					unreadNotifications = append(unreadNotifications, micNotifUniqueId)

					userNotifications[mu] = append(userNotifications[mu], [2]string{micNotifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), micNotifUniqueId)

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notifSnippet,
						})
					})
				}
			}

			// batch processing
			if err := cache.StoreNewComments(ctx, newComments); err != nil {
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

			for parentCommentId, commentId_score_Pairs := range commentComments {
				eg.Go(func() error {
					parentCommentId, commentId_score_Pairs := parentCommentId, commentId_score_Pairs

					if err := cache.StoreCommentComments(sharedCtx, parentCommentId, commentId_score_Pairs); err != nil {
						return err
					}

					go func() {
						ctx := context.Background()
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

					return nil
				})
			}

			for user, notifId_stmsgId_Pairs := range userNotifications {
				eg.Go(func() error {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					return cache.StoreUserNotifications(sharedCtx, user, notifId_stmsgId_Pairs)
				})
			}

			for _, fn := range newCommentDBExtrasFuncs {
				eg.Go(func() error {
					fn := fn

					return fn()
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
