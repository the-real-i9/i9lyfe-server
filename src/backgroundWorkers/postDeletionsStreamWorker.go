package backgroundWorkers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
)

func postDeletionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "post_deletions"
		groupName    = "post_deletion_listeners"
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
			var msgs []eventTypes.PostDeletionEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.PostDeletionEvent

				msg.OwnerUser = stmsg.Values["ownerUser"].(string)
				msg.PostId = stmsg.Values["postId"].(string)
				msg.Mentions = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["mentions"].(string))

				msgs = append(msgs, msg)

			}

			_ = len(msgs)

			/* DO WHAT'S NEEDED */

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
