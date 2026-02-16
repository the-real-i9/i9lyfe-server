package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func repostsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "reposts"
		groupName    = "repost_listeners"
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
			var msgs []eventTypes.RepostEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.RepostEvent

				msg.ReposterUser = stmsg.Values["reposterUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostOwner = stmsg.Values["postOwner"].(string)
				msg.RepostId = stmsg.Values["repostId"].(string)
				msg.RepostData = stmsg.Values["repostData"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))
				msg.RepostCursor = helpers.FromJson[int64](stmsg.Values["repostCursor"].(string))

				msgs = append(msgs, msg)

			}

			reposts := []string{}

			postReposts := make(map[string][]any)

			userRepostedPosts := make(map[string][][2]any)

			userPosts := make(map[string][][2]any)

			userFeedPosts := make(map[string][][2]any)

			notifications := []string{}
			unreadNotifications := []any{}

			userNotifications := make(map[string][][2]any)

			sendNotifEventMsgFuncs := []func(){}

			fanOutPostFuncs := []func(){}

			// batch data for batch processing
			for _, msg := range msgs {
				reposts = append(reposts, msg.RepostId, msg.RepostData)

				postReposts[msg.PostId] = append(postReposts[msg.PostId], msg.RepostId)

				userRepostedPosts[msg.ReposterUser] = append(userRepostedPosts[msg.ReposterUser], [2]any{msg.PostId, float64(msg.RepostCursor)})

				if msg.ReposterUser == msg.PostOwner {
					continue
				}

				userPosts[msg.ReposterUser] = append(userPosts[msg.ReposterUser], [2]any{msg.RepostId, float64(msg.RepostCursor)})

				userFeedPosts[msg.ReposterUser] = append(userFeedPosts[msg.ReposterUser], [2]any{msg.PostId, float64(msg.RepostCursor)})

				notifUniqueId := fmt.Sprintf("user_%s_reposted_post_%s", msg.ReposterUser, msg.PostId)
				notif := helpers.BuildNotification(notifUniqueId, "repost", msg.At, map[string]any{
					"reposted_post_id": msg.PostId,
					"reposter_user":    msg.ReposterUser,
					"repost_id":        msg.RepostId,
				})

				notifications = append(notifications, notifUniqueId, helpers.ToJson(notif))
				unreadNotifications = append(unreadNotifications, notifUniqueId)

				userNotifications[msg.PostOwner] = append(userNotifications[msg.PostOwner], [2]any{notifUniqueId, float64(msg.RepostCursor)})

				sendNotifEventMsgFuncs = append(sendNotifEventMsgFuncs, func() {
					notifSnippet, _ := modelHelpers.BuildNotifSnippetUIFromCache(context.Background(), notifUniqueId)

					realtimeService.SendEventMsg(msg.PostOwner, appTypes.ServerEventMsg{
						Event: "new notification",
						Data:  notifSnippet,
					})
				})

				fanOutPostFuncs = append(fanOutPostFuncs, func() {
					go contentRecommendationService.FanOutPostToFollowers(msg.RepostId, float64(msg.RepostCursor), msg.ReposterUser)
				})
			}

			// batch processing
			if err := cache.StoreNewPosts(ctx, reposts); err != nil {
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

			for postId, repostIds := range postReposts {
				eg.Go(func() error {
					postId, repostIds := postId, repostIds

					if err := cache.StorePostReposts(sharedCtx, postId, repostIds); err != nil {
						return err
					}

					go func() {
						ctx := context.Background()
						for postId := range postReposts {
							totalRepostsCount, err := cache.GetPostRepostsCount(ctx, postId)
							if err != nil {
								continue
							}

							realtimeService.PublishPostMetric(ctx, map[string]any{
								"post_id":              postId,
								"latest_reposts_count": totalRepostsCount,
							})
						}
					}()

					return nil
				})
			}

			for user, postId_score_Pairs := range userPosts {
				eg.Go(func() error {
					user, postId_score_Pairs := user, postId_score_Pairs

					return cache.StoreUserPosts(sharedCtx, user, postId_score_Pairs)
				})
			}

			for user, postId_score_Pairs := range userFeedPosts {
				eg.Go(func() error {
					user, postId_score_Pairs := user, postId_score_Pairs

					if err := cache.StoreUserFeedPosts(sharedCtx, user, postId_score_Pairs); err != nil {
						return err
					}

					return nil
				})
			}

			for user, postId_score_Pairs := range userRepostedPosts {
				eg.Go(func() error {
					user, postId_score_Pairs := user, postId_score_Pairs

					return cache.StoreUserRepostedPosts(sharedCtx, user, postId_score_Pairs)
				})
			}

			for user, notifId_score_Pairs := range userNotifications {
				eg.Go(func() error {
					user, notifId_score_Pairs := user, notifId_score_Pairs

					return cache.StoreUserNotifications(sharedCtx, user, notifId_score_Pairs)
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
