package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
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
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for postId, users := range postReactionsRemoved {
					cache.RemovePostReactions(pipe, ctx, postId, users)
				}

				for postId, rUsers := range postReactorsRemoved {
					cache.RemovePostReactors(pipe, ctx, postId, rUsers)
				}

				for user, postIds := range userReactionRemovedPosts {
					cache.RemoveUserReactedPosts(pipe, ctx, user, postIds)
				}

				for postId, rUsers := range postReactorsRemoved {
					cache.RemovePostReactors(pipe, ctx, postId, rUsers)
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
					for postId := range postReactionsRemoved {
						postId_IntCmd[postId] = cache.GetPostReactionsCount(pipe, ctx, postId)
					}

					return nil
				})
				if err != nil && err != redis.Nil {
					helpers.LogError(err)
					return
				}

				for postId, lc := range postId_IntCmd {
					latestCount, err := lc.Result()
					if err != nil {
						continue
					}

					realtimeService.PublishPostMetric(ctx, map[string]any{
						"post_id":                postId,
						"latest_reactions_count": latestCount,
					})
				}
			}()

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
