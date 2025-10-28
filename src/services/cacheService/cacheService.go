package cacheService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb = appGlobals.RedisClient

func StoreUserFollowers(ctx context.Context, user string, followerUser ...any) error {
	if err := rdb.SAdd(ctx, fmt.Sprintf("user:%s:followers", user), followerUser...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserFollowing(ctx context.Context, user string, followingUser ...any) error {
	if err := rdb.SAdd(ctx, fmt.Sprintf("user:%s:following", user), followingUser...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserMentionedPosts(ctx context.Context, mentionedUser string, postId ...any) error {
	if err := rdb.SAdd(ctx, fmt.Sprintf("user:%s:mentioned_posts", mentionedUser), postId...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserReactedPosts() {

}

func StoreUserNotifications(ctx context.Context, user string, notifId ...any) error {
	members := []redis.Z{}
	for i, nid := range notifId {
		members = append(members, redis.Z{
			Score:  float64(time.Now().Unix()) + float64(i),
			Member: nid,
		})
	}
	if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:notifications:%d-%d", user, time.Now().Year(), time.Now().Month()), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewPosts(ctx context.Context, posts map[string]any) error {
	if err := rdb.MSet(ctx, posts).Err(); err != nil {
		helpers.LogError(err)
		return nil
	}

	return nil
}

func StorePostReactions() {
	// key: post_id
	// value: {reactions:[]map = []map{reactor:map, emoji:string}, total_reactions:int}
}

func StorePostComments() {
	// key: post_id
	// {reactions:[]map = []map{reactor:map, emoji:string}, total_reactions:int}
}

func StoreCommentComments() {

}

func StoreHashtagPosts(ctx context.Context, hashtag string, postId ...any) error {
	if err := rdb.SAdd(ctx, fmt.Sprintf("hastag:%s:posts", hashtag), postId...).Err(); err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func StoreNewNotifications(ctx context.Context, notifications map[string]map[string]any) error {
	for k, v := range notifications {
		if err := rdb.HSet(ctx, fmt.Sprintf("notification:%s", k), v).Err(); err != nil {
			helpers.LogError(err)

			return err
		}
	}

	return nil
}
