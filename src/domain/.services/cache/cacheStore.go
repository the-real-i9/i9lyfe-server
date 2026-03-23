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

func StoreOfflineUsers(pipe redis.Pipeliner, ctx context.Context, user_lastSeen_Pairs map[string]int64) {
	members := []redis.Z{}
	membersUnsorted := []any{}

	for user, lastSeen := range user_lastSeen_Pairs {

		members = append(members, redis.Z{
			Score:  float64(lastSeen),
			Member: user,
		})

		membersUnsorted = append(membersUnsorted, user)
	}

	pipe.ZAdd(ctx, "offline_users", members...)
	pipe.SAdd(ctx, "offline_users_unsorted", membersUnsorted...)
}

func StoreUserFeedPosts(pipe redis.Pipeliner, ctx context.Context, user string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:feed", user), members...)
}

func StoreUserFollowers(pipe redis.Pipeliner, ctx context.Context, followingUser string, followerUser_score_Pair [][2]any) {
	members := []redis.Z{}
	for _, pair := range followerUser_score_Pair {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: user,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:followers", followingUser), members...)
}

func StoreUserFollowings(pipe redis.Pipeliner, ctx context.Context, followerUser string, followingUser_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range followingUser_score_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: user,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:followings", followerUser), members...)
}

func StoreUserPosts(pipe redis.Pipeliner, ctx context.Context, user string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:posts", user), members...)
}

func StoreUserMentionedPosts(pipe redis.Pipeliner, ctx context.Context, mentionedUser string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:mentioned_posts", mentionedUser), members...)
}

func StoreUserReactedPosts(pipe redis.Pipeliner, ctx context.Context, reactorUser string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:reacted_posts", reactorUser), members...)
}

func StoreUserCommentedPosts(pipe redis.Pipeliner, ctx context.Context, commenterUser string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:commented_posts", commenterUser), members...)
}

func StoreUserRepostedPosts(pipe redis.Pipeliner, ctx context.Context, user string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:reposted_posts", user), members...)
}

func StoreUserNotifications(pipe redis.Pipeliner, ctx context.Context, user string, notifId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range notifId_score_Pairs {
		notifId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: notifId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:notifications:%d-%d", user, time.Now().Year(), time.Now().Month()), members...)
}

func StoreUserSavedPosts(pipe redis.Pipeliner, ctx context.Context, saverUser string, postId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range postId_score_Pairs {
		postId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: postId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:saved_posts", saverUser), members...)
}

func StorePostReactions(pipe redis.Pipeliner, ctx context.Context, postId string, userWithEmojiPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), userWithEmojiPairs)
}

func StorePostReactors(pipe redis.Pipeliner, ctx context.Context, postId string, reactorUser_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range reactorUser_score_Pairs {
		rUser := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: rUser,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("reacted_post:%s:reactors", postId), members...)
}

func StoreCommentReactions(pipe redis.Pipeliner, ctx context.Context, commentId string, userWithEmojiPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), userWithEmojiPairs)
}

func StoreCommentReactors(pipe redis.Pipeliner, ctx context.Context, commentId string, reactorUser_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range reactorUser_score_Pairs {
		rUser := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: rUser,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("reacted_comment:%s:reactors", commentId), members...)
}

func StorePostComments(pipe redis.Pipeliner, ctx context.Context, postId string, commentId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range commentId_score_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: commentId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("commented_post:%s:comments", postId), members...)
}

func StorePostSaves(pipe redis.Pipeliner, ctx context.Context, postId string, saverUsers []any) {
	pipe.SAdd(ctx, fmt.Sprintf("saved_post:%s:saves", postId), saverUsers...)
}

func StorePostReposts(pipe redis.Pipeliner, ctx context.Context, postId string, repostIds []any) {
	pipe.SAdd(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId), repostIds...)
}

func StoreCommentComments(pipe redis.Pipeliner, ctx context.Context, parentCommentId string, commentId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range commentId_score_Pairs {
		commentId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: commentId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("commented_comment:%s:comments", parentCommentId), members...)
}

func StoreNewPosts(pipe redis.Pipeliner, ctx context.Context, newPosts []string) {
	pipe.HSet(ctx, "posts", newPosts)
}

func StoreNewComments(pipe redis.Pipeliner, ctx context.Context, newComments []string) {
	pipe.HSet(ctx, "comments", newComments)
}

func StoreNewNotifications(pipe redis.Pipeliner, ctx context.Context, newNotifs []string) {
	pipe.HSet(ctx, "notifications", newNotifs)
}

func StoreUnreadNotifications(pipe redis.Pipeliner, ctx context.Context, unreadNotifs []any) {
	pipe.SAdd(ctx, "unread_notifications", unreadNotifs...)
}

func StoreUserChats(pipe redis.Pipeliner, ctx context.Context, ownerUser string, partnerUserWithChatInfoPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), partnerUserWithChatInfoPairs)
}

func StoreUserChatUnreadMsgs(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, unreadMsgs []any) {
	pipe.SAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:unread_messages", ownerUser, partnerUser), unreadMsgs...)
}

func StoreUserChatsSorted(pipe redis.Pipeliner, ctx context.Context, ownerUser string, partnerUser_score_Pairs map[string]float64) {
	members := []redis.Z{}
	for partnerUser, score := range partnerUser_score_Pairs {

		members = append(members, redis.Z{
			Score:  score,
			Member: partnerUser,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:chats_sorted", ownerUser), members...)
}

func StoreChatHistoryEntries(ctx context.Context, newCHEs []string) error {
	if err := rdb().HSet(ctx, "chat_history_entries", newCHEs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, CHEId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range CHEId_score_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: CHEId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", ownerUser, partnerUser), members...)
	pipe.ZAdd(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", partnerUser, ownerUser), members...)
}

func StoreMsgReactions(pipe redis.Pipeliner, ctx context.Context, msgId string, userWithEmojiPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("message:%s:reactions", msgId), userWithEmojiPairs)
}
