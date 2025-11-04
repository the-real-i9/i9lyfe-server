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
		// cache new post and publish to subscribers
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

			var msgs []eventTypes.PostReactionEvent
			helpers.ToStruct(stmsgValues, &msgs)

			msgsLen := len(msgs)

			postReactions := make(map[string]map[string][2]any)

			userReactedPosts := make(map[string][][2]string)

			notifications := make(map[string][2]any)

			userNotifications := make(map[string][][2]string)

			sendNotifEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for i, msg := range msgs {
				postReactions[msg.PostId][msg.ReactorUser] = [2]any{msg.ReactionData, stmsgIds[i]}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_post_%s", msg.ReactorUser, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_post", msg.At, map[string]any{
					"to_post_id":   msg.PostId,
					"reactor_user": msg.ReactorUser,
					"emoji":        msg.ReactionData["emoji"],
				})

				notifications[notifUniqueId] = [2]any{notif, stmsgIds[i]}

				userNotifications[msg.PostId] = append(userNotifications[msg.PostOwner], [2]string{notifUniqueId, stmsgIds[i]})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notif["notif"] = helpers.Json2Map(notif["notif"].(string))

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notif,
					})
				})

				userReactedPosts[msg.ReactorUser] = append(userReactedPosts[msg.ReactorUser], [2]string{msg.PostId, stmsgIds[i]})
			}

			// batch processing
			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			for postId, userTo_Reaction_stmsgId_Pair := range postReactions {
				userToReactionDataMap := make(map[string]any)
				stmsgIds := []string{}

				for user, reactionData_stmsgId_Pair := range userTo_Reaction_stmsgId_Pair {
					userToReactionDataMap[user] = reactionData_stmsgId_Pair[0].(map[string]any)
					stmsgIds = append(stmsgIds, reactionData_stmsgId_Pair[1].(string))
				}

				wg.Go(func() {
					if err := cacheService.StorePostReactions(ctx, postId, userToReactionDataMap); err != nil {
						for _, id := range stmsgIds {
							failedStreamMsgIds[id] = true
						}
					}
				})
			}

			wg.Wait()

			for notifId, notif_stmsgId_Pair := range notifications {
				wg.Go(func() {
					notifId, notifData, stmsgId := notifId, notif_stmsgId_Pair[0], notif_stmsgId_Pair[1].(string)

					if err := cacheService.StoreNewNotifications(ctx, notifId, notifData); err != nil {
						failedStreamMsgIds[stmsgId] = true
					}
				})
			}

			go func() {
				for postId := range postReactions {
					totalRxnsCount, err := rdb.HLen(ctx, fmt.Sprintf("post:%s:reactions", postId)).Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":                postId,
						"latest_reactions_count": totalRxnsCount,
					})
				}
			}()

			for k, postId_stmsgId_Pairs := range userReactedPosts {
				wg.Go(func() {
					k, postId_stmsgId_Pairs := k, postId_stmsgId_Pairs

					if err := cacheService.StoreUserReactedPosts(ctx, k, postId_stmsgId_Pairs); err != nil {
						for _, d := range postId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for k, notifId_stmsgId_Pairs := range userNotifications {
				wg.Go(func() {
					k, notifId_stmsgId_Pairs := k, notifId_stmsgId_Pairs

					err = cacheService.StoreUserNotifications(ctx, k, notifId_stmsgId_Pairs)
					if err != nil {
						for _, d := range notifId_stmsgId_Pairs {
							failedStreamMsgIds[d[1]] = true
						}
					}
				})
			}

			for _, fn := range sendNotifEventMsgFuncs {
				fn()
			}

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
