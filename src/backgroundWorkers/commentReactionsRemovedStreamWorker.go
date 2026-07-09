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

				msg.CommentId = stmsg.Values["commentId"].(string)

				msgs = append(msgs, msg)

			}

			commentReactionsRemoved := make(map[string]int)

			// batch data for batch processing
			for _, msg := range msgs {
				commentReactionsRemoved[msg.CommentId]++
			}

			// batch processing
			sqls := []string{}
			params := [][]any{}

			for commentId, rc := range commentReactionsRemoved {

				sqls = append(
					sqls,
					/* sql */ `
					UPDATE public.comments SET reactions_count = reactions_count - $2 WHERE comment_id = $1
					RETURNING comment_id, reactions_count AS rxns_count
					`,
				)
				params = append(params, []any{commentId, rc})
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
				CommentId string `db:"comment_id"`
				RxnsCount int    `db:"rxns_count"`
			}

			commentIdRxns, err := pgDB.BatchQueryTx[res](ctx, pgTx, sqls, params)
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
				for _, cr := range commentIdRxns {
					pubsubService.PublishPostMetric(context.Background(), map[string]any{
						"comment_id":             cr.CommentId,
						"latest_reactions_count": cr.RxnsCount,
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
