package backgroundWorkers

import (
	"i9lyfe/src/appGlobals"
)

func Start() error {
	newPostsStreamBgWorker(appGlobals.RedisClient)
	// postReactionsStreamBgWorker(appGlobals.RedisClient)
	// go postCommentBgTasks()

	return nil
}
