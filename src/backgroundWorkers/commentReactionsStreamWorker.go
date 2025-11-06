package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cacheService"
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
			var stmsgValues []map[string]any

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)
				stmsgValues = append(stmsgValues, stmsg.Values)

			}

			var msgs []eventTypes.CommentReactionEvent
			helpers.ToStruct(stmsgValues, &msgs)

			msgsLen := len(msgs)

			commentReactions := make(map[string][][2]any)

			notifications := make(map[any]any, msgsLen)

			userNotifications := make(map[string][][2]string, msgsLen)

			sendNotifEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for i, msg := range msgs {
				commentReactions[msg.CommentId] = append(commentReactions[msg.CommentId], [2]any{[]string{msg.ReactorUser, msg.Emoji}, stmsgIds[i]})

				if msg.CommentOwner == msg.ReactorUser {
					continue
				}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_comment_%s", msg.ReactorUser, msg.CommentId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_comment", msg.At, map[string]any{
					"to_comment_id": msg.CommentId,
					"reactor_user":  msg.ReactorUser,
					"emoji":         msg.Emoji,
				})

				notifications[notifUniqueId] = helpers.ToJson(notif)

				userNotifications[msg.CommentOwner] = append(userNotifications[msg.CommentOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["is_read"] = false

					realtimeService.SendEventMsg(msg.CommentOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  helpers.ToJson(notif),
					})
				})
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			if err := cacheService.StoreNewNotifications(ctx, notifications); err != nil {
				return
			}

			for commentId, userWithEmoji_stmsgId_Pairs := range commentReactions {
				wg.Go(func() {
					commentId, userWithEmoji_stmsgId_Pairs := commentId, userWithEmoji_stmsgId_Pairs

					userWithEmojiPairs := [][]string{}

					for _, userWithEmoji_stmsgId_Pair := range userWithEmoji_stmsgId_Pairs {
						userWithEmojiPairs = append(userWithEmojiPairs, userWithEmoji_stmsgId_Pair[0].([]string))
					}

					if err := cacheService.StorePostReactions(ctx, commentId, slices.Concat(userWithEmojiPairs...)); err != nil {
						for _, d := range userWithEmoji_stmsgId_Pairs {
							failedStreamMsgIds[d[1].(string)] = true
						}
					}
				})
			}

			wg.Wait()

			go func() {
				for commentId := range commentReactions {
					totalRxnsCount, err := rdb.HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId)).Result()
					if err != nil {
						continue
					}

					realtimeService.PublishCommentMetric(ctx, map[string]any{
						"comment_id":             commentId,
						"latest_reactions_count": totalRxnsCount,
					})
				}
			}()

			for user, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					user, notifId_stmsgId_Pairs := user, notifId_stmsgId_Pairs

					err = cacheService.StoreUserNotifications(ctx, user, notifId_stmsgId_Pairs)
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
