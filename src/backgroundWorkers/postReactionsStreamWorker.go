package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cacheService"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

func postReactionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_reactions"
		groupName    = "post_reactions_listeners"
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
	readLoop:
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

			postReactions := make(map[string]map[string]any) // key: %s: postId > field: %s: username | value: %s: emoji

			postReactorsWithReaction := make(map[string][]any) // key: %s-%s: postId-emoji

			userReactedPosts := make(map[string][]any) // key: %s: username > [...postId]

			notifications := make(map[string]map[string]any)

			userNotifications := make(map[string][]any)

			postReactionExtrasFuncs := make([]func() error, msgsLen)

			sendEventMsgFuncs := make([]func(), msgsLen)

			// batch data for batch processing
			for _, msg := range msgs {
				msg := msg
				postsMSet[fmt.Sprintf("post:%s", msg.PostId)] = msg.PostData

				for _, ht := range msg.Hashtags {
					ht := ht
					hashtagPosts[ht] = append(hashtagPosts[ht], msg.PostId)
				}

				for _, mu := range msg.Mentions {
					mu := mu
					userMentionedPosts[mu] = append(userMentionedPosts[mu], msg.PostId)

					notifUniqueId := fmt.Sprintf("user_%s_mentioned_in_post_%s", mu, msg.PostId)
					notif := helpers.BuildPostMentionNotification(notifUniqueId, msg.PostId, msg.OwnerUser, msg.At)

					notifications[notifUniqueId] = notif

					userNotifications[mu] = append(userNotifications[mu], notifUniqueId)

					sendEventMsgFuncs = append(sendEventMsgFuncs, func() {
						notif["notif"] = helpers.Json2Map(notif["notif"].(string))

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notif,
						})
					})
				}

				newPostExtrasFuncs = append(newPostExtrasFuncs, func() error {
					return postModel.NewPostExtras(ctx, msg.OwnerUser, msg.PostId, msg.Mentions, msg.Hashtags, msg.At)
				})

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					contentRecommendationService.FanOutPost(msg.PostId)
				})
			}

			// batch processing

			// store new posts
			if err := cacheService.StoreNewPosts(ctx, postsMSet); err != nil {
				continue
			}

			if err := cacheService.StoreNewNotifications(ctx, notifications); err != nil {
				continue
			}

			wg := new(sync.WaitGroup)
			errs := make(chan error)

			wg.Go(func() {
				for k, v := range hashtagPosts {
					if err := cacheService.StoreHashtagPosts(ctx, k, v...); err != nil {
						errs <- err
					}
				}
			})

			wg.Go(func() {
				for k, v := range userMentionedPosts {
					if err := cacheService.StoreUserMentionedPosts(ctx, k, v...); err != nil {
						errs <- err
					}
				}
			})

			wg.Go(func() {
				for k, v := range userNotifications {
					if err := cacheService.StoreUserNotifications(ctx, k, v...); err != nil {
						errs <- err
					}
				}
			})

			wg.Go(func() {
				for _, fn := range newPostExtrasFuncs {
					if err := fn(); err != nil {
						errs <- err
					}
				}
			})

			wg.Go(func() {
				for _, fn := range fanOutPostFuncs {
					fn()
				}

				for _, fn := range sendEventMsgFuncs {
					fn()
				}
			})

			wg.Wait()

			for range errs {
				continue readLoop
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
