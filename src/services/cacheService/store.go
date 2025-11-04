package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
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

func StoreNewPosts(ctx context.Context, postId string, postData any) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("post:%s", postId), postData).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostReactions(ctx context.Context, postId string, userToReactionDataMap map[string]any) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), userToReactionDataMap).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StorePostComments() {
	// key: post_id
	// {reactions:[]map = []map{reactor:map, emoji:string}, total_reactions:int}
}

func StoreCommentComments() {

}

func StoreNewNotifications(ctx context.Context, notifId string, notifData any) error {
	if err := rdb.HSet(ctx, fmt.Sprintf("notification:%s", notifId), notifData).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
