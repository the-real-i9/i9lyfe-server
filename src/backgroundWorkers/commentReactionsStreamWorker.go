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

			commentReactions := make(map[string][]string)

			// having comment reactors separate, allows us to
			// paginate through the list of reactions on a comment
			commentReactors := make(map[string][][2]string)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]string, msgsLen)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for i, msg := range msgs {
				commentReactions[msg.CommentId] = append(commentReactions[msg.CommentId], msg.ReactorUser.Username, msg.Emoji)

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
			if len(notifications) > 0 {
				if err := cache.StoreNewNotifications(ctx, notifications); err != nil {
					return
				}

				if err := cache.StoreUnreadNotifications(ctx, unreadNotifications); err != nil {
					return
				}
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for commentId, userWithEmojiPairs := range commentReactions {
				eg.Go(func() error {
					commentId, userWithEmojiPairs := commentId, userWithEmojiPairs

					if err := cache.StoreCommentReactions(sharedCtx, commentId, userWithEmojiPairs); err != nil {
						return err
					}

					go func() {
						ctx := context.Background()
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

					return nil
				})
			}

			for commentId, rUser_stmsgId_Pairs := range commentReactors {
				eg.Go(func() error {
					commentId, rUser_stmsgId_Pairs := commentId, rUser_stmsgId_Pairs

					return cache.StoreCommentReactors(sharedCtx, commentId, rUser_stmsgId_Pairs)
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
