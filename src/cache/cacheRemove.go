package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

func RemoveOfflineUsers(ctx context.Context, users []any) error {
	if err := rdb().ZRem(ctx, "offline_users", users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemovePostReactions(pipe redis.Pipeliner, ctx context.Context, postId string, reactorUsers []string) {
	pipe.HDel(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), reactorUsers...)
}

func RemovePostReactors(pipe redis.Pipeliner, ctx context.Context, postId string, reactorUsers []any) {
	pipe.ZRem(ctx, fmt.Sprintf("reacted_post:%s:reactors", postId), reactorUsers...)
}

func RemovePostSaves(pipe redis.Pipeliner, ctx context.Context, postId string, users []any) {
	pipe.SRem(ctx, fmt.Sprintf("saved_post:%s:saves", postId), users...)
}

func RemoveCommentReactions(pipe redis.Pipeliner, ctx context.Context, commentId string, reactorUsers []string) {
	pipe.HDel(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), reactorUsers...)
}

func RemoveCommentReactors(pipe redis.Pipeliner, ctx context.Context, commentId string, reactorUsers []any) {
	pipe.ZRem(ctx, fmt.Sprintf("reacted_comment:%s:reactors", commentId), reactorUsers...)
}

func RemoveUserReactedPosts(pipe redis.Pipeliner, ctx context.Context, user string, postIds []any) {
	pipe.ZRem(ctx, fmt.Sprintf("user:%s:reacted_posts", user), postIds...)
}

func RemoveUserSavedPosts(pipe redis.Pipeliner, ctx context.Context, user string, postIds []any) {
	pipe.ZRem(ctx, fmt.Sprintf("user:%s:saved_posts", user), postIds...)
}

func RemoveChatHistoryEntries(ctx context.Context, CHEIds []string) error {
	if err := rdb().HDel(ctx, "chat_history_entries", CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, CHEIds []any) {
	pipe.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", ownerUser, partnerUser), CHEIds...)
	pipe.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", partnerUser, ownerUser), CHEIds...)
}

func RemoveUserFollowers(ctx context.Context, followingUser string, followerUsers []any) error {
	if err := rdb().ZRem(ctx, fmt.Sprintf("user:%s:followers", followingUser), followerUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserFollowings(ctx context.Context, followerUser string, followingUsers []any) error {
	if err := rdb().ZRem(ctx, fmt.Sprintf("user:%s:followings", followerUser), followingUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveMsgReactions(pipe redis.Pipeliner, ctx context.Context, msgId string, reactorUsers []string) {
	pipe.HDel(ctx, fmt.Sprintf("message:%s:reactions", msgId), reactorUsers...)
}

func RemoveUnreadMessages(ctx context.Context, readMessages []string) error {
	if err := rdb().HDel(ctx, "unread_messages", readMessages...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUnreadNotifications(ctx context.Context, readNotifications ...any) error {
	if err := rdb().SRem(ctx, "unread_notifications", readNotifications...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserChatUnreadMsgs(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, readMsgs []any) {
	pipe.SRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:unread_messages", ownerUser, partnerUser), readMsgs...)
}
