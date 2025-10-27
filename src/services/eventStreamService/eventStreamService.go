package eventStreamService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"

	"github.com/redis/go-redis/v9"
)

var rdb = appGlobals.RedisClient

func QueueNewPost(ctx context.Context, npe eventTypes.NewPostEvent) {
	var npeMap map[string]any

	helpers.StructToMap(npe, &npeMap)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "new_posts",
		Values: npe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReaction( /* { post_id, reactor, emoji } */ ) {

}
