package backgroundWorkers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/services/pubsubService"

	"i9lyfe/src/types/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
)

func postCommentsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_comments"
		groupName    = "post_comment_listeners"
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
			var msgs []eventTypes.PostCommentEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostCommentEvent

				msg.PostId = stmsg.Values["postId"].(string)

				msgs = append(msgs, msg)

			}

			postComments := make(map[string]int)

			// batch data for batch processing
			for _, msg := range msgs {
				postComments[msg.PostId]++
			}

			// batch processing
			sqls := []string{}
			params := [][]any{}

			for postId, cc := range postComments {

				sqls = append(
					sqls,
					/* sql */ `
					UPDATE posts SET comments_count = comments_count + $2 WHERE id_ = $1
					RETURNING id_ AS post_id, comments_count
					`,
				)
				params = append(params, []any{postId, cc})
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
				PostId        string `db:"post_id"`
				CommentsCount int    `db:"comments_count"`
			}

			postIdComments, err := pgDB.BatchQueryTypeTx[res](ctx, pgTx, sqls, params)
			if err != nil {
				helpers.LogError(err)
				return
			}

			err = pgTx.Commit(ctx)
			if err != nil {
				helpers.LogError(err)
				return
			}

			go func() {
				for _, pc := range postIdComments {
					pubsubService.PublishPostMetric(context.Background(), map[string]any{
						"post_id":               pc.PostId,
						"latest_comments_count": pc.CommentsCount,
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
