package backgroundWorkers

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cacheService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"
	"slices"
	"sync"

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
				Count:    500,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var stmsgValues []map[string]any

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)
				stmsgValues = append(stmsgValues, stmsg.Values)

			}

			var msgs []eventTypes.UserFollowEvent
			helpers.ToStruct(stmsgValues, &msgs)

			userFollowingsRemoved := make(map[string][][2]string)
			userFollowersRemoved := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {

				userFollowingsRemoved[msg.FollowerUser] = append(userFollowingsRemoved[msg.FollowerUser], [2]string{msg.FollowingUser, stmsgIds[i]})

				userFollowersRemoved[msg.FollowingUser] = append(userFollowersRemoved[msg.FollowingUser], [2]string{msg.FollowerUser, stmsgIds[i]})

			}

			wg := new(sync.WaitGroup)

			failedStreamMsgIds := make(map[string]bool)

			// batch processing

			for followerUser, followingUser_stmsgId_Pairs := range userFollowingsRemoved {
				wg.Go(func() {
					followerUser, followingUser_stmsgId_Pairs := followerUser, followingUser_stmsgId_Pairs

					followingUsers := []any{}

					for _, pair := range followingUser_stmsgId_Pairs {
						followingUsers = append(followingUsers, pair[0])
					}

					if err := cacheService.RemoveUserFollowings(ctx, followerUser, followingUsers); err != nil {
						for _, pair := range followingUser_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
					}
				})
			}

			for followingUser, followerUser_stmsgId_Pairs := range userFollowersRemoved {
				wg.Go(func() {
					followingUser, followerUser_stmsgId_Pairs := followingUser, followerUser_stmsgId_Pairs

					followerUsers := []any{}

					for _, pair := range followerUser_stmsgId_Pairs {
						followerUsers = append(followerUsers, pair[0])
					}

					if err := cacheService.RemoveUserFollowers(ctx, followingUser, followerUsers); err != nil {
						for _, pair := range followerUser_stmsgId_Pairs {
							failedStreamMsgIds[pair[1]] = true
						}
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
