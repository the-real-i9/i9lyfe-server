package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"time"

	"github.com/redis/go-redis/v9"
)

func StoreNewUsers(ctx context.Context, newUsers []string) error {
	if err := rdb().HSet(ctx, "users", newUsers).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreOfflineUsers(ctx context.Context, user_lastSeen_Pairs map[string]int64) error {
	members := []redis.Z{}
	membersUnsorted := []any{}

	for user, lastSeen := range user_lastSeen_Pairs {

		members = append(members, redis.Z{
			Score:  float64(lastSeen),
			Member: user,
		})

		membersUnsorted = append(membersUnsorted, user)
	}

	if err := rdb().ZAdd(ctx, "offline_users", members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb().SAdd(ctx, "offline_users_unsorted", membersUnsorted...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserFeedPosts(ctx context.Context, user string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:feed", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserFollowers(ctx context.Context, followingUser string, followerUser_score_Pair [][2]any) error {
	members := []redis.Z{}
	for _, pair := range followerUser_score_Pair {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: user,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:followers", followingUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserFollowings(ctx context.Context, followerUser string, followingUser_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range followingUser_score_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: user,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:followings", followerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserPosts(ctx context.Context, user string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:posts", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserMentionedPosts(ctx context.Context, mentionedUser string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:mentioned_posts", mentionedUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserReactedPosts(ctx context.Context, reactorUser string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:reacted_posts", reactorUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserCommentedPosts(ctx context.Context, commenterUser string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:commented_posts", commenterUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserRepostedPosts(ctx context.Context, user string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:reposted_posts", user), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserNotifications(ctx context.Context, user string, notifId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range notifId_score_Pairs {
		notifId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: notifId,
		})
	}
	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:notifications:%d-%d", user, time.Now().Year(), time.Now().Month()), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserSavedPosts(ctx context.Context, saverUser string, postId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:saved_posts", saverUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostReactions(ctx context.Context, postId string, userWithEmojiPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostReactors(ctx context.Context, postId string, reactorUser_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range reactorUser_score_Pairs {
		rUser := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: rUser,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("reacted_post:%s:reactors", postId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreCommentReactions(ctx context.Context, commentId string, userWithEmojiPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreCommentReactors(ctx context.Context, commentId string, reactorUser_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range reactorUser_score_Pairs {
		rUser := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: rUser,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("reacted_comment:%s:reactors", commentId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostComments(ctx context.Context, postId string, commentId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range commentId_score_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: commentId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("commented_post:%s:comments", postId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostSaves(ctx context.Context, postId string, saverUsers []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("saved_post:%s:saves", postId), saverUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostReposts(ctx context.Context, postId string, repostIds []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId), repostIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreCommentComments(ctx context.Context, parentCommentId string, commentId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range commentId_score_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: commentId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("commented_comment:%s:comments", parentCommentId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewPosts(ctx context.Context, newPosts []string) error {
	if err := rdb().HSet(ctx, "posts", newPosts).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewComments(ctx context.Context, newComments []string) error {
	if err := rdb().HSet(ctx, "comments", newComments).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewNotifications(ctx context.Context, newNotifs []string) error {
	if err := rdb().HSet(ctx, "notifications", newNotifs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUnreadNotifications(ctx context.Context, unreadNotifs []any) error {
	if err := rdb().SAdd(ctx, "unread_notifications", unreadNotifs...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChats(ctx context.Context, ownerUser string, partnerUserWithChatInfoPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), partnerUserWithChatInfoPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatUnreadMsgs(ctx context.Context, ownerUser, partnerUser string, unreadMsgs []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:unread_messages", ownerUser, partnerUser), unreadMsgs...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatsSorted(ctx context.Context, ownerUser string, partnerUser_score_Pairs map[string]float64) error {
	members := []redis.Z{}
	for partnerUser, score := range partnerUser_score_Pairs {

		members = append(members, redis.Z{
			Score:  score,
			Member: partnerUser,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:chats_sorted", ownerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreChatHistoryEntries(ctx context.Context, newCHEs []string) error {
	if err := rdb().HSet(ctx, "chat_history_entries", newCHEs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatHistory(ctx context.Context, ownerUser, partnerUser string, CHEId_score_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range CHEId_score_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: CHEId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", ownerUser, partnerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", partnerUser, ownerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreMsgReactions(ctx context.Context, msgId string, userWithEmojiPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("message:%s:reactions", msgId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
