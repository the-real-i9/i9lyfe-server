package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"
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

				msg.CommenterUser = stmsg.Values["commenterUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentData = stmsg.Values["commentData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.CommentCursor = helpers.ParseInt(stmsg.Values["commentCursor"].(string))

				msgs = append(msgs, msg)

			}

			newComments := []string{}

			postComments := make(map[string][][2]any)

			userCommentedPosts := make(map[string][][2]any)

			newCommentDBExtrasFuncs := []func(context.Context) error{}

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {
				newComments = append(newComments, msg.CommentId, msg.CommentData)

				postComments[msg.PostId] = append(postComments[msg.PostId], [2]any{msg.CommentId, float64(msg.CommentCursor)})

				userCommentedPosts[msg.CommenterUser] = append(userCommentedPosts[msg.CommenterUser], [2]any{msg.PostId, float64(msg.CommentCursor)})

				newCommentDBExtrasFuncs = append(newCommentDBExtrasFuncs, func(ctx context.Context) error {
					return postModel.CommentOnExtras(ctx, msg.CommentId, msg.Mentions)
				})

				if msg.PostOwner != msg.CommenterUser {

					copNotifUniqueId := fmt.Sprintf("user_%s_comment_%s_on_post_%s", msg.CommenterUser, msg.CommentId, msg.PostId)
					copNotif := helpers.BuildNotification(copNotifUniqueId, "comment_on_post", msg.At, map[string]any{
						"on_post_id":     msg.PostId,
						"commenter_user": msg.CommenterUser,
						"comment_id":     msg.CommentId,
					})

					notifications = append(notifications, copNotifUniqueId, helpers.ToMsgPack(copNotif))
					unreadNotifications = append(unreadNotifications, copNotifUniqueId)

					userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]any{copNotifUniqueId, float64(msg.CommentCursor)})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), copNotifUniqueId)

						realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
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

					notifications = append(notifications, micNotifUniqueId, helpers.ToMsgPack(micNotif))
					unreadNotifications = append(unreadNotifications, micNotifUniqueId)

					userNotifications[mu] = append(userNotifications[mu], [2]any{micNotifUniqueId, float64(msg.CommentCursor)})

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
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				cache.StoreNewComments(pipe, ctx, newComments)

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

				for postId, commentId_score_Pairs := range postComments {
					cache.StorePostComments(pipe, ctx, postId, commentId_score_Pairs)
				}

				for user, postId_score_Pairs := range userCommentedPosts {
					cache.StoreUserCommentedPosts(pipe, ctx, user, postId_score_Pairs)
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

			go func() {
				ctx := context.Background()
				postId_IntCmd := make(map[string]*redis.IntCmd)

				_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					for postId := range postComments {
						postId_IntCmd[postId] = cache.GetPostCommentsCount(pipe, ctx, postId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for postId, tcc := range postId_IntCmd {
					totalCommentsCount, err := tcc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":               postId,
						"latest_comments_count": totalCommentsCount,
					})
				}
			}()

			eg, sharedCtx := errgroup.WithContext(ctx)

			for _, fn := range newCommentDBExtrasFuncs {
				eg.Go(func() error {
					fn := fn

					return fn(sharedCtx)
				})
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
