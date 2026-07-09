package eventStreamService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"i9lyfe/src/types/eventTypes"

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

func QueueUserFollowEvent(ctx context.Context, ufe eventTypes.UserFollowEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "users_followed",
		Values: ufe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueUserUnfollowEvent(ctx context.Context, uue eventTypes.UserUnfollowEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "users_unfollowed",
		Values: uue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueNewPostEvent(ctx context.Context, npe eventTypes.NewPostEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_posts",
		Values: npe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostReactionEvent(ctx context.Context, pre eventTypes.PostReactionEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions",
		Values: pre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostReactionRemovedEvent(ctx context.Context, rpre eventTypes.PostReactionRemovedEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_reactions_removed",
		Values: rpre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostCommentEvent(ctx context.Context, pce eventTypes.PostCommentEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments",
		Values: pce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostCommentRemovedEvent(ctx context.Context, rpce eventTypes.PostCommentRemovedEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_comments_removed",
		Values: rpce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueCommentReactionEvent(ctx context.Context, cre eventTypes.CommentReactionEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions",
		Values: cre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueCommentReactionRemovedEvent(ctx context.Context, rcre eventTypes.CommentReactionRemovedEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_reactions_removed",
		Values: rcre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueCommentCommentEvent(ctx context.Context, cce eventTypes.CommentCommentEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments",
		Values: cce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueCommentCommentRemovedEvent(ctx context.Context, rcce eventTypes.CommentCommentRemovedEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "comment_comments_removed",
		Values: rcce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueueRepostEvent(ctx context.Context, re eventTypes.RepostEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "reposts",
		Values: re,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostSaveEvent(ctx context.Context, pse eventTypes.PostSaveEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_saves",
		Values: pse,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func QueuePostUnsaveEvent(ctx context.Context, pue eventTypes.PostUnsaveEvent) error {
	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "post_unsaves",
		Values: pue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
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
