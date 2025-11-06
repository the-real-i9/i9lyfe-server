package backgroundWorkers

import (
	"github.com/redis/go-redis/v9"
)

// These workers allows for mad scalability. Thundering herds are a piece of cake.
//
// The only error situation we can have in production is when our Redis server is down.
// In this case, we manually XPENDING + XCLAIM all the messages that have been moved
// to the Redis Pending List when the server issue has been fixed.
// This is, in fact, a rare case.
//
// Courtesy of Redis Streams
func Start(rdc *redis.Client) error {
	newPostsStreamBgWorker(rdc)

	postReactionsStreamBgWorker(rdc)
	postReactionRemovedStreamBgWorker(rdc)

	postCommentsStreamBgWorker(rdc)

	commentReactionsStreamBgWorker(rdc)
	commentReactionRemovedStreamBgWorker(rdc)

	commentCommentsStreamBgWorker(rdc)

	repostsStreamBgWorker(rdc)

	postSavesStreamBgWorker(rdc)

	return nil
}
