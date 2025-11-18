package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func postCommentsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_comments"
		groupName    = "post_comment_listeners"
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
			var msgs []eventTypes.PostCommentEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostCommentEvent

				msg.CommenterUser = helpers.FromJson[appTypes.ClientUser](stmsg.Values["commenterUser"].(string))
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentData = stmsg.Values["commentData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			newComments := []string{}

			postComments := make(map[string][][2]string)

			userCommentedPosts := make(map[string][][2]string)

			newCommentDBExtrasFuncs := []func() error{}

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				newComments = append(newComments, msg.CommentId, msg.CommentData)

				postComments[msg.PostId] = append(postComments[msg.PostId], [2]string{msg.CommentId, stmsgIds[i]})

				userCommentedPosts[msg.CommenterUser.Username] = append(userCommentedPosts[msg.CommenterUser.Username], [2]string{msg.PostId, stmsgIds[i]})

				newCommentDBExtrasFuncs = append(newCommentDBExtrasFuncs, func() error {
					return postModel.CommentOnExtras(ctx, msg.CommentId, msg.Mentions)
				})

				if msg.PostOwner != msg.CommenterUser.Username {

					copNotifUniqueId := fmt.Sprintf("user_%s_comment_%s_on_post_%s", msg.CommenterUser.Username, msg.CommentId, msg.PostId)
					copNotif := helpers.BuildNotification(copNotifUniqueId, "comment_on_post", msg.At, map[string]any{
						"on_post_id":     msg.PostId,
						"commenter_user": msg.CommenterUser.Username,
						"comment_id":     msg.CommentId,
					})

					notifications = append(notifications, copNotifUniqueId, helpers.ToJson(copNotif))
					unreadNotifications = append(unreadNotifications, copNotifUniqueId)

					userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]string{copNotifUniqueId, stmsgIds[i]})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						copNotif["unread"] = true
						copNotif["details"].(map[string]any)["commenter_user"] = msg.CommenterUser

						realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  copNotif,
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

			for postId, commentId_stmsgId_Pairs := range postComments {
				eg.Go(func() error {
					postId, commentId_stmsgId_Pairs := postId, commentId_stmsgId_Pairs

					return cache.StorePostComments(sharedCtx, postId, commentId_stmsgId_Pairs)
				})
			}

			if eg.Wait() != nil {
				return
			}

			go func() {
				for postId := range postComments {
					totalCommentsCount, err := cache.GetPostCommentsCount(sharedCtx, postId)
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(sharedCtx, map[string]any{
						"post_id":               postId,
						"latest_comments_count": totalCommentsCount,
					})
				}
			}()

			for user, postId_stmsgId_Pairs := range userCommentedPosts {
				eg.Go(func() error {
					user, postId_stmsgId_Pairs := user, postId_stmsgId_Pairs

					return cache.StoreUserCommentedPosts(sharedCtx, user, postId_stmsgId_Pairs)
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

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
