package backgroundWorkers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/types/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
)

func usersUnfollowedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "users_unfollowed"
		groupName    = "user_unfollowed_listeners"
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
			var msgs []eventTypes.UserUnfollowEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.UserUnfollowEvent

				msg.FollowerUser = stmsg.Values["followerUser"].(string)
				msg.FollowingUser = stmsg.Values["followingUser"].(string)

				msgs = append(msgs, msg)
			}

			userFollowingsRemoved := make(map[string]int)
			userFollowersRemoved := make(map[string]int)

			// batch data for batch processing
			for _, msg := range msgs {
				userFollowingsRemoved[msg.FollowerUser]++

				userFollowersRemoved[msg.FollowingUser]++
			}

			// batch processing
			sqls := []string{}
			params := [][]any{}

			for username, fc := range userFollowingsRemoved {

				sqls = append(sqls /* sql */, `UPDATE users SET followings_count = followings_count - $2 WHERE username = $1`)
				params = append(params, []any{username, fc})
			}

			for username, fc := range userFollowersRemoved {

				sqls = append(sqls /* sql */, `UPDATE users SET followers_count = followers_count - $2 WHERE username = $1`)
				params = append(params, []any{username, fc})
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

			err = pgDB.BatchExecTx(ctx, pgTx, sqls, params)
			if err != nil {
				helpers.LogError(err)
				return
			}

			err = pgTx.Commit(ctx)
			if err != nil {
				helpers.LogError(err)
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
