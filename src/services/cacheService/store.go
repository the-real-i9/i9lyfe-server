package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func stmsgIdToScore(val string) (res float64) {
	var err error
	if res, err = strconv.ParseFloat(strings.Replace(val, "-", ".", 1), 64); err != nil {
		helpers.LogError(err)
	}
	return
}

func StoreUserFollowers(ctx context.Context, user string, followerUser_stmsgId_Pair [][2]string) error {
	members := []redis.Z{}
	for _, pair := range followerUser_stmsgId_Pair {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: user,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:followers", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserFollowing(ctx context.Context, user string, followingUser_stmsgId_Pair [][2]string) error {
	members := []redis.Z{}
	for _, pair := range followingUser_stmsgId_Pair {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: user,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:following", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserPosts(ctx context.Context, user string, postId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range postId_stmsgId_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: postId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:posts", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserMentionedPosts(ctx context.Context, mentionedUser string, postId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range postId_stmsgId_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: postId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:mentioned_posts", mentionedUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserReactedPosts(ctx context.Context, reactorUser string, postId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range postId_stmsgId_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: postId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:reacted_posts", reactorUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserCommentedPosts(ctx context.Context, commenterUser string, postId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range postId_stmsgId_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: postId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:commented_posts", commenterUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserNotifications(ctx context.Context, user string, notifId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range notifId_stmsgId_Pairs {
		notifId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: notifId,
		})
	}
	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:notifications:%d-%d", user, time.Now().Year(), time.Now().Month()), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserSavedPosts(ctx context.Context, saverUser string, postId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range postId_stmsgId_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: postId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:saved_posts", saverUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostReactions(ctx context.Context, postId string, userWithEmojiPairs []string) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreCommentReactions(ctx context.Context, commentId string, userWithEmojiPairs []string) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostComments(ctx context.Context, postId string, commentId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range commentId_stmsgId_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: commentId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("post:%s:comments", postId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostSaves(ctx context.Context, postId string, saverUsers []any) error {
	if err := rdb.SAdd(ctx, fmt.Sprintf("saved_post:%s:saves", postId), saverUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreCommentComments(ctx context.Context, parentCommentId string, commentId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range commentId_stmsgId_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: commentId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("comment:%s:comments", parentCommentId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewPosts(ctx context.Context, newPosts map[string]any) error {
	if err := rdb.HSet(ctx, "posts", newPosts).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewComments(ctx context.Context, newComments map[string]any) error {
	if err := rdb.HSet(ctx, "comments", newComments).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewNotifications(ctx context.Context, newNotifs map[any]any) error {
	if err := rdb.HSet(ctx, "notifications", newNotifs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb.SAdd(ctx, "unread_notifications", slices.Collect(maps.Keys(newNotifs))...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreChatHistoryEntries(ctx context.Context, newCHEs map[string]any) error {
	if err := rdb.HSet(ctx, "chat_history_entries", newCHEs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatHistory(ctx context.Context, ownerUserPartnerUser [2]string, CHEId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range CHEId_stmsgId_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: CHEId,
		})
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s", ownerUserPartnerUser[0], ownerUserPartnerUser[1]), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb.ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s", ownerUserPartnerUser[1], ownerUserPartnerUser[0]), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUnreadMessages(ctx context.Context, unreadMessages []string) error {
	if err := rdb.HSet(ctx, "unread_messages", unreadMessages).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreMsgReactions(ctx context.Context, msgId string, userWithEmojiPairs []string) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("message:%s:reactions", msgId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
