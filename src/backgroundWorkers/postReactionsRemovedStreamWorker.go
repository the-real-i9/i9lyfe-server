package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func postReactionRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_reactions_removed"
		groupName    = "post_reaction_removed_listeners"
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
			var msgs []eventTypes.PostReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostReactionRemovedEvent

				msg.ReactorUser = stmsg.Values["reactorUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)

				msgs = append(msgs, msg)

			}

			postReactionsRemoved := make(map[string][]string)

			postReactorsRemoved := make(map[string][]any)

			userReactionRemovedPosts := make(map[string][]any)

			// batch data for batch processing
			for _, msg := range msgs {

				postReactionsRemoved[msg.PostId] = append(postReactionsRemoved[msg.PostId], msg.ReactorUser)
				// these two above and below follow a similar implemtation,
				// i.e. we can use postReactionsRemoved to remove post reactors too
				// but we're just separating concerns here
				postReactorsRemoved[msg.PostId] = append(postReactorsRemoved[msg.PostId], msg.ReactorUser)

				userReactionRemovedPosts[msg.ReactorUser] = append(userReactionRemovedPosts[msg.ReactorUser], msg.PostId)
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for postId, users := range postReactionsRemoved {
				eg.Go(func() error {
					postId, users := postId, users

					return cache.RemovePostReactions(sharedCtx, postId, users)
				})
			}

			if eg.Wait() != nil {
				return
			}

			for postId := range postReactionsRemoved {
				go func() {
					latestCount, err := cache.GetPostReactionsCount(sharedCtx, postId)
					if err != nil {
						return
					}

					realtimeService.PublishPostMetric(sharedCtx, map[string]any{
						"post_id":                postId,
						"latest_reactions_count": latestCount,
					})
				}()
			}

			for user, postIds := range userReactionRemovedPosts {
				eg.Go(func() error {
					user, postIds := user, postIds

					return cache.RemoveUserReactedPosts(sharedCtx, user, postIds)
				})
			}

			for postId, rUsers := range postReactorsRemoved {
				eg.Go(func() error {
					postId, rUsers := postId, rUsers

					return cache.RemovePostReactors(sharedCtx, postId, rUsers)
				})
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
