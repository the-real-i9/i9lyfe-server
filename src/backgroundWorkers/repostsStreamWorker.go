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
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.RepostCursor = helpers.ParseInt(stmsg.Values["repostCursor"].(string))

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

				notifications = append(notifications, notifUniqueId, helpers.ToMsgPack(notif))
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
					contentRecommendationService.FanOutPostToFollowers(msg.RepostId, float64(msg.RepostCursor), msg.ReposterUser)
				})
			}

			// batch processing
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				cache.StoreNewPosts(pipe, ctx, reposts)

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

				for postId, repostIds := range postReposts {
					cache.StorePostReposts(pipe, ctx, postId, repostIds)
				}

				for user, postId_score_Pairs := range userPosts {
					cache.StoreUserPosts(pipe, ctx, user, postId_score_Pairs)
				}

				for user, postId_score_Pairs := range userFeedPosts {
					cache.StoreUserFeedPosts(pipe, ctx, user, postId_score_Pairs)
				}

				for user, postId_score_Pairs := range userRepostedPosts {
					cache.StoreUserRepostedPosts(pipe, ctx, user, postId_score_Pairs)
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
					for postId := range postReposts {
						postId_IntCmd[postId] = cache.GetPostRepostsCount(pipe, ctx, postId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for postId, lc := range postId_IntCmd {
					totalRepostsCount, err := lc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":              postId,
						"latest_reposts_count": totalRepostsCount,
					})
				}
			}()

			for _, fn := range sendNotifEventMsgFuncs {
				go fn()
			}

			for _, fn := range fanOutPostFuncs {
				go fn()
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
