package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/contentRecommendationService"
	"time"

	"i9lyfe/src/types/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
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

				msg.OwnerUsername = stmsg.Values["ownerUsername"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.PostCursor = helpers.ParseInt(stmsg.Values["postCursor"].(string))

				msgs = append(msgs, msg)

			}

			userFeedPosts := make(map[string][][2]any)

			fanOutPostFuncs := []func(context.Context) error{}

			pc := float64(time.Now().UnixMicro())

			// batch data for batch processing
			for _, msg := range msgs {
				userFeedPosts[msg.OwnerUsername] = append(userFeedPosts[msg.OwnerUsername], [2]any{msg.PostId, float64(msg.PostCursor)})

				fanOutPostFuncs = append(fanOutPostFuncs, func(ctx context.Context) error {
					return contentRecommendationService.FanOutPostToFollowers(ctx, msg.PostId, pc, msg.OwnerUsername)
				})
			}

			// batch processing
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {

				for user, postId_score_Pairs := range userFeedPosts {
					cache.StoreUserFeedPosts(pipe, ctx, user, postId_score_Pairs)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			for _, f := range fanOutPostFuncs {
				if err := f(ctx); err != nil {
					return
				}
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
