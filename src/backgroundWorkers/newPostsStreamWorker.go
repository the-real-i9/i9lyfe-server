package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func newPostsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_posts"
		groupName    = "new_post_listeners"
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
			var msgs []eventTypes.NewPostEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewPostEvent

				msg.OwnerUser = stmsg.Values["ownerUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostData = stmsg.Values["postData"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))
				msg.Hashtags = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["hashtags"].(string))
				msg.PostCursor = helpers.ParseInt(stmsg.Values["postCursor"].(string))
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			newPosts := []string{}

			userPosts := make(map[string][][2]any)

			userFeedPosts := make(map[string][][2]any)

			userMentionedPosts := make(map[string][][2]any)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any)

			newPostDBExtrasFuncs := []func(context.Context) error{}

			sendNotifEventMsgFuncs := []func(){}

			fanOutPostFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {
				newPosts = append(newPosts, msg.PostId, msg.PostData)

				userPosts[msg.OwnerUser] = append(userPosts[msg.OwnerUser], [2]any{msg.PostId, float64(msg.PostCursor)})

				userFeedPosts[msg.OwnerUser] = append(userFeedPosts[msg.OwnerUser], [2]any{msg.PostId, float64(msg.PostCursor)})

				newPostDBExtrasFuncs = append(newPostDBExtrasFuncs, func(ctx context.Context) error {
					return postModel.NewPostExtras(ctx, msg.PostId, msg.Mentions, msg.Hashtags)
				})

				for _, mu := range msg.Mentions {
					userMentionedPosts[mu] = append(userMentionedPosts[mu], [2]any{msg.PostId, float64(msg.PostCursor)})

					if mu == msg.OwnerUser {
						continue
					}

					notifUniqueId := fmt.Sprintf("user_%s_mentioned_in_post_%s", mu, msg.PostId)
					notif := helpers.BuildNotification(notifUniqueId, "mention_in_post", msg.At, map[string]any{
						"in_post_id":      msg.PostId,
						"mentioning_user": msg.OwnerUser,
					})

					notifications = append(notifications, notifUniqueId, helpers.ToMsgPack(notif))
					unreadNotifications = append(unreadNotifications, notifUniqueId)

					userNotifications[mu] = append(userNotifications[mu], [2]any{notifUniqueId, float64(msg.PostCursor)})

					sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
						notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), notifUniqueId)

						realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
							Event: "new notification",
							Data:  notifSnippet,
						})
					})

					fanOutPostFuncs = append(fanOutPostFuncs, func() {
						contentRecommendationService.FanOutPostToFollowers(msg.PostId, float64(msg.PostCursor), msg.OwnerUser)
					})
				}
			}

			// batch processing
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				cache.StoreNewPosts(pipe, ctx, newPosts)

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

				for user, postId_score_Pairs := range userPosts {
					cache.StoreUserPosts(pipe, ctx, user, postId_score_Pairs)
				}

				for user, postId_score_Pairs := range userFeedPosts {
					cache.StoreUserFeedPosts(pipe, ctx, user, postId_score_Pairs)
				}

				for user, postId_score_Pairs := range userMentionedPosts {
					cache.StoreUserMentionedPosts(pipe, ctx, user, postId_score_Pairs)
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

			eg, sharedCtx := errgroup.WithContext(ctx)

			for _, fn := range newPostDBExtrasFuncs {
				eg.Go(func() error {
					fn := fn

					return fn(sharedCtx)
				})
			}

			for _, fn := range sendNotifEventMsgFuncs {
				go fn()
			}

			for _, fn := range fanOutPostFuncs {
				go fn()
			}

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
