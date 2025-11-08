package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
)

func RemovePostReactions(ctx context.Context, postId string, reactorUsers []string) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), reactorUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemovePostSaves(ctx context.Context, postId string, users []any) error {
	if err := rdb.SRem(ctx, fmt.Sprintf("saved_post:%s:saves", postId), users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveCommentReactions(ctx context.Context, commentId string, reactorUsers []string) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), reactorUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserReactedPosts(ctx context.Context, user string, postIds []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:reacted_posts", user), postIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserSavedPosts(ctx context.Context, user string, postIds []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:saved_posts", user), postIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveChatHistoryEntries(ctx context.Context, CHEIds []string) error {
	if err := rdb.HDel(ctx, "chat_history_entries", CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserChatHistory(ctx context.Context, ownerUserPartnerUser [2]string, CHEIds []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s", ownerUserPartnerUser[0], ownerUserPartnerUser[1]), CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s", ownerUserPartnerUser[1], ownerUserPartnerUser[0]), CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserFollowers(ctx context.Context, followingUser string, followerUsers []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:followers", followingUser), followerUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserFollowings(ctx context.Context, followerUser string, followingUsers []any) error {
	if err := rdb.ZRem(ctx, fmt.Sprintf("user:%s:following", followerUser), followingUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveMsgReactions(ctx context.Context, msgId string, reactorUsers []string) error {
	if err := rdb.HDel(ctx, fmt.Sprintf("message:%s:reactions", msgId), reactorUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUnreadMessages(ctx context.Context, readMessages []string) error {
	if err := rdb.HDel(ctx, "unread_messages", readMessages...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
