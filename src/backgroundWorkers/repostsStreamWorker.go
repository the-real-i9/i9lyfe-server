package backgroundWorkers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/pubsubService"
	"time"

	"i9lyfe/src/types/eventTypes"
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

				msg.PostId = stmsg.Values["postId"].(string)
				msg.ReposterUser = stmsg.Values["reposterUser"].(string)
				msg.RepostId = stmsg.Values["repostId"].(string)
				msg.RepostCursor = helpers.ParseInt(stmsg.Values["repostCursor"].(string))

				msgs = append(msgs, msg)

			}

			postReposts := make(map[string]int)

			userFeedPosts := make(map[string][][2]any)

			fanOutPostFuncs := []func(context.Context) error{}

			rpc := float64(time.Now().UnixMicro())

			// batch data for batch processing
			for _, msg := range msgs {

				postReposts[msg.PostId]++

				userFeedPosts[msg.ReposterUser] = append(userFeedPosts[msg.ReposterUser], [2]any{msg.RepostId, float64(msg.RepostCursor)})

				fanOutPostFuncs = append(fanOutPostFuncs, func(ctx context.Context) error {
					return contentRecommendationService.FanOutPostToFollowers(ctx, msg.RepostId, rpc, msg.ReposterUser)
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

			sqls := []string{}
			params := [][]any{}

			for postId, rc := range postReposts {

				sqls = append(
					sqls,
					/* sql */ `
					UPDATE posts SET reposts_count = reposts_count + $2 WHERE id_ = $1
					RETURNING id_ AS post_id, reposts_count
					`,
				)
				params = append(params, []any{postId, rc})
			}

			pgTx, err := appGlobals.DBPool.Begin(ctx)
			if err != nil {
				helpers.LogError(err)
				return
			}

			defer func() {
				if err != nil {
					helpers.LogError(pgTx.Rollback(ctx))
				}
			}()

			type res struct {
				PostId       string `db:"post_id"`
				RepostsCount int    `db:"reposts_count"`
			}

			postIdReposts, err := pgDB.BatchQueryTx[res](ctx, pgTx, sqls, params)
			if err != nil {
				helpers.LogError(err)
				return
			}

			for _, f := range fanOutPostFuncs {
				if err := f(ctx); err != nil {
					return
				}
			}

			err = pgTx.Commit(ctx)
			if err != nil {
				helpers.LogError(err)
				return
			}

			go func() {
				for _, pr := range postIdReposts {
					pubsubService.PublishPostMetric(context.Background(), map[string]any{
						"post_id":              pr.PostId,
						"latest_reposts_count": pr.RepostsCount,
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
