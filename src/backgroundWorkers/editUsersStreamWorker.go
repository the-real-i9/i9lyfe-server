package backgroundWorkers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"
	"slices"
	"sync"

	"github.com/redis/go-redis/v9"
)

func editUsersStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "edit_users"
		groupName    = "edit_user_listeners"
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
			var msgs []eventTypes.EditUserEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.EditUserEvent

				msg.Username = stmsg.Values["username"].(string)
				msg.UpdateKVMap = helpers.FromJson[appTypes.BinableMap](stmsg.Values["updateKVMap"].(string))

				msgs = append(msgs, msg)

			}

			editUsers := make(map[string][2]any, len(msgs))

			// batch data for batch processing
			for i, msg := range msgs {
				editUsers[msg.Username] = [2]any{map[string]any(msg.UpdateKVMap), stmsgIds[i]}
			}

			// batch processing

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			for user, updateKVMap_stmsgId_Pair := range editUsers {

				wg.Go(func() {
					user, updateKVMap_stmsgId_Pair := user, updateKVMap_stmsgId_Pair

					if err := cache.UpdateUser(ctx, user, updateKVMap_stmsgId_Pair[0].(map[string]any)); err != nil {
						failedStreamMsgIds[updateKVMap_stmsgId_Pair[1].(string)] = true
					}
				})
			}

			wg.Wait()

			stmsgIds = slices.DeleteFunc(stmsgIds, func(stmsgId string) bool {
				return failedStreamMsgIds[stmsgId]
			})

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
