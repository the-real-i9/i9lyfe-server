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
	"golang.org/x/sync/errgroup"
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

				msg.ReactorUser = stmsg.Values["reactorUser"].(string)
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.RxnCursor = helpers.ParseInt(stmsg.Values["rxnCursor"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			postReactions := make(map[string][]string)

			// having post reactors separate, allows us to
			// paginate through the list of reactions on a post
			postReactors := make(map[string][][2]any)

			userReactedPosts := make(map[string][][2]any)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any, msgsLen)

			sendNotifEventMsgFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {
				postReactions[msg.PostId] = append(postReactions[msg.PostId], msg.ReactorUser, msg.Emoji)

				postReactors[msg.PostId] = append(postReactors[msg.PostId], [2]any{msg.ReactorUser, float64(msg.RxnCursor)})

				userReactedPosts[msg.ReactorUser] = append(userReactedPosts[msg.ReactorUser], [2]any{msg.PostId, float64(msg.RxnCursor)})

				if msg.PostOwner == msg.ReactorUser {
					continue
				}

				notifUniqueId := fmt.Sprintf("user_%s_reaction_to_post_%s", msg.ReactorUser, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "reaction_to_post", msg.At, map[string]any{
					"to_post_id":   msg.PostId,
					"reactor_user": msg.ReactorUser,
					"emoji":        msg.Emoji,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToMsgPack(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]any{notifUniqueId, float64(msg.RxnCursor)})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), notifUniqueId)

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notifSnippet,
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

			for postId, userWithEmojiPairs := range postReactions {
				eg.Go(func() error {
					postId, userWithEmojiPairs := postId, userWithEmojiPairs

					if err := cache.StorePostReactions(sharedCtx, postId, userWithEmojiPairs); err != nil {
						return err
					}

					go func() {
						ctx := context.Background()
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

					return nil
				})
			}

			for user, postId_score_Pairs := range userReactedPosts {
				eg.Go(func() error {
					user, postId_score_Pairs := user, postId_score_Pairs

					return cache.StoreUserReactedPosts(sharedCtx, user, postId_score_Pairs)
				})
			}

			for postId, rUser_score_Pairs := range postReactors {
				eg.Go(func() error {
					postId, rUser_score_Pairs := postId, rUser_score_Pairs

					return cache.StorePostReactors(sharedCtx, postId, rUser_score_Pairs)
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
