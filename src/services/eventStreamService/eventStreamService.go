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

func QueueUserFollowEvent(ufe eventTypes.UserFollowEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "users_followed",
		Values: ufe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueUserUnfollowEvent(uue eventTypes.UserUnfollowEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "users_unfollowed",
		Values: uue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewPostEvent(npe eventTypes.NewPostEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "new_posts",
		Values: npe,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionEvent(pre eventTypes.PostReactionEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_reactions",
		Values: pre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostReactionRemovedEvent(rpre eventTypes.PostReactionRemovedEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_reactions_removed",
		Values: rpre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentEvent(pce eventTypes.PostCommentEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_comments",
		Values: pce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostCommentRemovedEvent(rpce eventTypes.PostCommentRemovedEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_comments_removed",
		Values: rpce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionEvent(cre eventTypes.CommentReactionEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "comment_reactions",
		Values: cre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentReactionRemovedEvent(rcre eventTypes.CommentReactionRemovedEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "comment_reactions_removed",
		Values: rcre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentEvent(cce eventTypes.CommentCommentEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "comment_comments",
		Values: cce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueCommentCommentRemovedEvent(rcce eventTypes.CommentCommentRemovedEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "comment_comments_removed",
		Values: rcce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueRepostEvent(re eventTypes.RepostEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "reposts",
		Values: re,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostSaveEvent(pse eventTypes.PostSaveEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_saves",
		Values: pse,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueuePostUnsaveEvent(pue eventTypes.PostUnsaveEvent) {
	err := rdb().XAdd(context.Background(), &redis.XAddArgs{
		Stream: "post_unsaves",
		Values: pue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}
