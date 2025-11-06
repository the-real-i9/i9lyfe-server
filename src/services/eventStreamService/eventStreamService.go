package eventStreamService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"

	"github.com/redis/go-redis/v9"
)

var rdb = appGlobals.RedisClient

func QueueNewPostEvent(npe eventTypes.NewPostEvent) {
	ctx := context.Background()

	npeMap := helpers.StructToMap(npe)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "new_posts",
		Values: npeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostDeletionEvent(pde eventTypes.PostDeletionEvent) {
	ctx := context.Background()

	pdeMap := helpers.StructToMap(pde)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_deletions",
		Values: pdeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionEvent(pre eventTypes.PostReactionEvent) {
	ctx := context.Background()
	preMap := helpers.StructToMap(pre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions",
		Values: preMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionRemovedEvent(rpre eventTypes.PostReactionRemovedEvent) {
	ctx := context.Background()
	rpreMap := helpers.StructToMap(rpre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions_removed",
		Values: rpreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentEvent(pce eventTypes.PostCommentEvent) {
	ctx := context.Background()
	pceMap := helpers.StructToMap(pce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments",
		Values: pceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentRemovedEvent(rpce eventTypes.PostCommentRemovedEvent) {
	ctx := context.Background()
	rpceMap := helpers.StructToMap(rpce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments_removed",
		Values: rpceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionEvent(cre eventTypes.CommentReactionEvent) {
	ctx := context.Background()
	creMap := helpers.StructToMap(cre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions",
		Values: creMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionRemovedEvent(rcre eventTypes.CommentReactionRemovedEvent) {
	ctx := context.Background()
	rcreMap := helpers.StructToMap(rcre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions_removed",
		Values: rcreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentEvent(cce eventTypes.CommentCommentEvent) {
	ctx := context.Background()
	cceMap := helpers.StructToMap(cce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments",
		Values: cceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentRemovedEvent(rcce eventTypes.CommentCommentRemovedEvent) {
	ctx := context.Background()
	rcceMap := helpers.StructToMap(rcce)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments_removed",
		Values: rcceMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRepostEvent(re eventTypes.RepostEvent) {
	ctx := context.Background()
	reMap := helpers.StructToMap(re)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "reposts",
		Values: reMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostSaveEvent(pse eventTypes.PostSaveEvent) {
	ctx := context.Background()
	pseMap := helpers.StructToMap(pse)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_saves",
		Values: pseMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostUnsaveEvent(pue eventTypes.PostUnsaveEvent) {
	ctx := context.Background()
	pueMap := helpers.StructToMap(pue)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "post_unsaves",
		Values: pueMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewMessageEvent(nme eventTypes.NewMessageEvent) {
	ctx := context.Background()
	nmeMap := helpers.StructToMap(nme)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "new_messages",
		Values: nmeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewMsgReactionEvent(nmre eventTypes.NewMsgReactionEvent) {
	ctx := context.Background()
	nmreMap := helpers.StructToMap(nmre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_reactions",
		Values: nmreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgAckEvent(mae eventTypes.MsgAckEvent) {
	ctx := context.Background()
	maeMap := helpers.StructToMap(mae)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_acks",
		Values: maeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgDeletionEvent(mde eventTypes.MsgDeletionEvent) {
	ctx := context.Background()
	mdeMap := helpers.StructToMap(mde)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_deletions",
		Values: mdeMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgReactionRemovedEvent(mrre eventTypes.MsgReactionRemovedEvent) {
	ctx := context.Background()
	mrreMap := helpers.StructToMap(mrre)

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_deletions",
		Values: mrreMap,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}
