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

func commentCommentsRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "comment_comments_removed"
		groupName    = "comment_comment_removed_listeners"
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
				Count:    1000,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var msgs []eventTypes.CommentCommentRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.CommentCommentRemovedEvent

				msg.ParentCommentId = stmsg.Values["parentCommentId"].(string)

				msgs = append(msgs, msg)
			}

			parentCommentComments := make(map[string]int)

			// batch data for batch processing
			for _, msg := range msgs {
				parentCommentComments[msg.ParentCommentId]++
			}

			// batch processing
			sqls := []string{}
			params := [][]any{}

			for parentCommentId, cc := range parentCommentComments {

				sqls = append(
					sqls,
					/* sql */ `
					UPDATE public.comments SET comments_count = comments_count - $2 WHERE parent_comment_id = $1
					RETURNING parent_comment_id, comments_count
					`,
				)
				params = append(params, []any{parentCommentId, cc})
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
				ParentCommentId string `db:"parent_comment_id"`
				CommentsCount   int    `db:"comments_count"`
			}

			parentCommentIdComments, err := pgDB.BatchQueryTx[res](ctx, pgTx, sqls, params)
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
				for _, pc := range parentCommentIdComments {
					pubsubService.PublishPostMetric(context.Background(), map[string]any{
						"comment_id":            pc.ParentCommentId,
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
