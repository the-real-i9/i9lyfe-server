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

				msg.ReactorUser = stmsg.Values["reactorUser"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)
				msg.CommentOwner = stmsg.Values["commentOwner"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.RxnCursor = helpers.ParseInt(stmsg.Values["rxnCursor"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			commentReactions := make(map[string][]string)

			// having comment reactors separate, allows us to
			// paginate through the list of reactions on a comment
			commentReactors := make(map[string][][2]any)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any, msgsLen)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {
				commentReactions[msg.CommentId] = append(commentReactions[msg.CommentId], msg.ReactorUser, msg.Emoji)

				commentReactors[msg.CommentId] = append(commentReactors[msg.CommentId], [2]any{msg.ReactorUser, float64(msg.RxnCursor)})

				if msg.CommentOwner == msg.ReactorUser {
					continue
				}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_comment_%s", msg.ReactorUser, msg.CommentId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_comment", msg.At, map[string]any{
					"to_comment_id": msg.CommentId,
					"reactor_user":  msg.ReactorUser,
					"emoji":         msg.Emoji,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToMsgPack(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.CommentOwner] = append(userNotifications[msg.CommentOwner], [2]any{notifUniqueId, float64(msg.RxnCursor)})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), notifUniqueId)

					realtimeService.SendEventMsg(msg.CommentOwner, appTypes.ServerEventMsg{
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
				for commentId, userWithEmojiPairs := range commentReactions {
					cache.StoreCommentReactions(pipe, ctx, commentId, userWithEmojiPairs)
				}

				for commentId, rUser_score_Pairs := range commentReactors {
					cache.StoreCommentReactors(pipe, ctx, commentId, rUser_score_Pairs)
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
				commentId_IntCmd := make(map[string]*redis.IntCmd)

				_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					for commentId := range commentReactions {
						commentId_IntCmd[commentId] = cache.GetCommentReactionsCount(pipe, ctx, commentId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for commentId, lc := range commentId_IntCmd {
					latestCount, err := lc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishCommentMetric(ctx, map[string]any{
						"comment_id":             commentId,
						"latest_reactions_count": latestCount,
					})
				}
			}()

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
