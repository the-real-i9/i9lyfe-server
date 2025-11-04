package eventStreamService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"

	"github.com/redis/go-redis/v9"
)

var rdb = appGlobals.RedisClient

func QueueNewPostEvent(ctx context.Context, npe eventTypes.NewPostEvent) {
	npeMap := helpers.StructToMap(npe)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "new_posts",
		Values: npeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionEvent(ctx context.Context, pre eventTypes.PostReactionEvent) {
	preMap := helpers.StructToMap(pre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions",
		Values: preMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRemovePostReactionEvent(ctx context.Context, rpre eventTypes.RemovePostReactionEvent) {
	rpreMap := helpers.StructToMap(rpre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reaction_removal",
		Values: rpreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentEvent(ctx context.Context, pce eventTypes.PostCommentEvent) {
	pceMap := helpers.StructToMap(pce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments",
		Values: pceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRemovePostCommentEvent(ctx context.Context, rpce eventTypes.RemovePostCommentEvent) {
	rpceMap := helpers.StructToMap(rpce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comment_removal",
		Values: rpceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionEvent(ctx context.Context, cre eventTypes.CommentReactionEvent) {
	creMap := helpers.StructToMap(cre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions",
		Values: creMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRemoveCommentReactionEvent(ctx context.Context, rcre eventTypes.RemoveCommentReactionEvent) {
	rcreMap := helpers.StructToMap(rcre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reaction_removal",
		Values: rcreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentEvent(ctx context.Context, cce eventTypes.CommentCommentEvent) {
	cceMap := helpers.StructToMap(cce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments",
		Values: cceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRemoveCommentCommentEvent(ctx context.Context, rcce eventTypes.RemoveCommentCommentEvent) {
	rcceMap := helpers.StructToMap(rcce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comment_removal",
		Values: rcceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRepostEvent(ctx context.Context, re eventTypes.RepostEvent) {
	reMap := helpers.StructToMap(re)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "reposts",
		Values: reMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostSaveEvent(ctx context.Context, pse eventTypes.PostSaveEvent) {
	pseMap := helpers.StructToMap(pse)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_saves",
		Values: pseMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostUnsaveEvent(ctx context.Context, pue eventTypes.PostUnsaveEvent) {
	pueMap := helpers.StructToMap(pue)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_unsaves",
		Values: pueMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewMessageEvent(ctx context.Context, nme eventTypes.NewMessageEvent) {
	nmeMap := helpers.StructToMap(nme)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "new_messages",
		Values: nmeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}
