package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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
			var msgs []eventTypes.UserUnfollowEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.UserUnfollowEvent

				msg.FollowerUser = stmsg.Values["followerUser"].(string)
				msg.FollowingUser = stmsg.Values["followingUser"].(string)

				msgs = append(msgs, msg)

			}

			userFollowingsRemoved := make(map[string][]any)
			userFollowersRemoved := make(map[string][]any)

			// batch data for batch processing
			for _, msg := range msgs {

				userFollowingsRemoved[msg.FollowerUser] = append(userFollowingsRemoved[msg.FollowerUser], msg.FollowingUser)

				userFollowersRemoved[msg.FollowingUser] = append(userFollowersRemoved[msg.FollowingUser], msg.FollowerUser)

			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for followerUser, followingUsers := range userFollowingsRemoved {
				eg.Go(func() error {
					followerUser, followingUsers := followerUser, followingUsers

					return cache.RemoveUserFollowings(sharedCtx, followerUser, followingUsers)
				})
			}

			for followingUser, followerUsers := range userFollowersRemoved {
				eg.Go(func() error {
					followingUser, followerUsers := followingUser, followerUsers

					return cache.RemoveUserFollowers(sharedCtx, followingUser, followerUsers)
				})
			}

			if eg.Wait() != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
