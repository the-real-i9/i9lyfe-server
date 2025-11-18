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

func commentReactionRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "comment_reactions_removed"
		groupName    = "comment_reaction_removed_listeners"
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
			var msgs []eventTypes.CommentReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)
				var msg eventTypes.CommentReactionRemovedEvent

				msg.ReactorUser = stmsg.Values["reactorUser"].(string)
				msg.CommentId = stmsg.Values["commentId"].(string)

				msgs = append(msgs, msg)

			}

			commentReactionsRemoved := make(map[string][]string)

			commentReactorsRemoved := make(map[string][]any)

			// batch data for batch processing
			for _, msg := range msgs {

				commentReactionsRemoved[msg.CommentId] = append(commentReactionsRemoved[msg.CommentId], msg.ReactorUser)
				// these two above and below follow a similar implemtation,
				// i.e. we can use commentReactionsRemoved to remove comment reactors too
				// but we're just separating concerns here
				commentReactorsRemoved[msg.CommentId] = append(commentReactorsRemoved[msg.CommentId], msg.ReactorUser)
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for commentId, users := range commentReactionsRemoved {

				eg.Go(func() error {
					commentId, users := commentId, users

					return cache.RemoveCommentReactions(sharedCtx, commentId, users)
				})
			}

			if eg.Wait() != nil {
				return
			}

			for commentId := range commentReactionsRemoved {
				go func() {
					latestCount, err := cache.GetCommentReactionsCount(sharedCtx, commentId)
					if err != nil {
						return
					}

					realtimeService.PublishCommentMetric(sharedCtx, map[string]any{
						"comment_id":             commentId,
						"latest_reactions_count": latestCount,
					})
				}()
			}

			for commentId, rUsers := range commentReactorsRemoved {
				eg.Go(func() error {
					commentId, rUsers := commentId, rUsers

					return cache.RemoveCommentReactors(sharedCtx, commentId, rUsers)
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
