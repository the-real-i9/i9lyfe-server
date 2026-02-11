package eventStreamService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func QueueNewUserEvent(nue eventTypes.NewUserEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_users",
		Values: nue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueEditUserEvent(eue eventTypes.EditUserEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "edit_users",
		Values: eue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueUserPresenceChangeEvent(upce eventTypes.UserPresenceChangeEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "user_presence_changes",
		Values: upce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueUserFollowEvent(ufe eventTypes.UserFollowEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "users_followed",
		Values: ufe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueUserUnfollowEvent(uue eventTypes.UserUnfollowEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "users_unfollowed",
		Values: uue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewPostEvent(npe eventTypes.NewPostEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_posts",
		Values: npe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostDeletionEvent(pde eventTypes.PostDeletionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_deletions",
		Values: pde,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionEvent(pre eventTypes.PostReactionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions",
		Values: pre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionRemovedEvent(rpre eventTypes.PostReactionRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions_removed",
		Values: rpre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentEvent(pce eventTypes.PostCommentEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments",
		Values: pce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentRemovedEvent(rpce eventTypes.PostCommentRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments_removed",
		Values: rpce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionEvent(cre eventTypes.CommentReactionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions",
		Values: cre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionRemovedEvent(rcre eventTypes.CommentReactionRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions_removed",
		Values: rcre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentEvent(cce eventTypes.CommentCommentEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments",
		Values: cce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentRemovedEvent(rcce eventTypes.CommentCommentRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments_removed",
		Values: rcce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRepostEvent(re eventTypes.RepostEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "reposts",
		Values: re,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostSaveEvent(pse eventTypes.PostSaveEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_saves",
		Values: pse,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostUnsaveEvent(pue eventTypes.PostUnsaveEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_unsaves",
		Values: pue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewMessageEvent(nme eventTypes.NewMessageEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_messages",
		Values: nme,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewMsgReactionEvent(nmre eventTypes.NewMsgReactionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_reactions",
		Values: nmre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgsAckEvent(mae eventTypes.MsgsAckEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "msgs_acks",
		Values: mae,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgDeletionEvent(mde eventTypes.MsgDeletionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_deletions",
		Values: mde,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueMsgReactionRemovedEvent(mrre eventTypes.MsgReactionRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "msg_reactions_removed",
		Values: mrre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}
