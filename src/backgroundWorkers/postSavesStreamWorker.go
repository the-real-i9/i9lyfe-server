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

func postSavesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_saves"
		groupName    = "post_save_listeners"
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
			var msgs []eventTypes.PostSaveEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostSaveEvent

				msg.PostId = stmsg.Values["postId"].(string)

				msgs = append(msgs, msg)

			}

			postSaves := make(map[string]int)

			// batch data for batch processing
			for _, msg := range msgs {
				postSaves[msg.PostId]++
			}

			// batch processing
			sqls := []string{}
			params := [][]any{}

			for postId, sc := range postSaves {

				sqls = append(
					sqls,
					/* sql */ `
					UPDATE posts SET saves_count = saves_count + $2 WHERE id_ = $1
					RETURNING id_ AS post_id, saves_count
					`,
				)
				params = append(params, []any{postId, sc})
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
				PostId     string `db:"post_id"`
				SavesCount int    `db:"saves_count"`
			}

			postIdSaves, err := pgDB.BatchQueryTypeTx[res](ctx, pgTx, sqls, params)
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
				for _, pr := range postIdSaves {
					pubsubService.PublishPostMetric(context.Background(), map[string]any{
						"post_id":            pr.PostId,
						"latest_saves_count": pr.SavesCount,
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
