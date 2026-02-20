package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

type postUIData struct {
	OwnerUser      UITypes.ContentOwnerUser
	ReposterUser   UITypes.ClientUser
	ReactionsCount int64
	CommentsCount  int64
	RepostsCount   int64
	SavesCount     int64
	MeReaction     string
	MeSaved        bool
	MeReposted     bool
}

func GetPostUIData(ctx context.Context, postId, ownerUsername, reposterUsername, clientUsername string) (postUIData, error) {
	var (
		pud                                                     postUIData
		ownerUser, reposterUser, meReaction                     *redis.StringCmd
		reactionsCount, commentsCount, repostsCount, savesCount *redis.IntCmd
		meReposted, meSaved                                     *redis.FloatCmd
	)

	_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		ownerUser = pipe.HGet(ctx, "users", ownerUsername)
		reposterUser = pipe.HGet(ctx, "users", reposterUsername)
		reactionsCount = pipe.HLen(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId))
		commentsCount = pipe.ZCard(ctx, fmt.Sprintf("commented_post:%s:comments", postId))
		repostsCount = pipe.SCard(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId))
		savesCount = pipe.SCard(ctx, fmt.Sprintf("saved_post:%s:saves", postId))
		meReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), clientUsername)
		meReposted = pipe.ZScore(ctx, fmt.Sprintf("user:%s:reposted_posts", clientUsername), postId)
		meSaved = pipe.ZScore(ctx, fmt.Sprintf("user:%s:saved_posts", clientUsername), postId)

		return nil
	})
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	userMsgPack, err := ownerUser.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}
	pud.OwnerUser = helpers.FromMsgPack[UITypes.ContentOwnerUser](userMsgPack)

	userMsgPack, err = reposterUser.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}
	pud.ReposterUser = helpers.FromMsgPack[UITypes.ClientUser](userMsgPack)

	pud.ReactionsCount, err = reactionsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	pud.CommentsCount, err = commentsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	pud.RepostsCount, err = repostsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	pud.SavesCount, err = savesCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	pud.MeReaction, err = meReaction.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	_, err = meReposted.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}
	pud.MeReposted = err == nil

	_, err = meSaved.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return postUIData{}, err
	}
	pud.MeSaved = err == nil

	return pud, nil
}

type commentUIData struct {
	OwnerUser      UITypes.ContentOwnerUser
	ReactionsCount int64
	CommentsCount  int64
	MeReaction     string
}

func GetCommentUIData(ctx context.Context, commentId, ownerUsername, clientUsername string) (commentUIData, error) {

	var (
		cud                           commentUIData
		ownerUser, meReaction         *redis.StringCmd
		reactionsCount, commentsCount *redis.IntCmd
	)

	_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		ownerUser = pipe.HGet(ctx, "users", ownerUsername)
		reactionsCount = pipe.HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId))
		commentsCount = pipe.ZCard(ctx, fmt.Sprintf("commented_comment:%s:comments", commentId))
		meReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), clientUsername)

		return nil
	})
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}

	userMsgPack, err := ownerUser.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}
	cud.OwnerUser = helpers.FromMsgPack[UITypes.ContentOwnerUser](userMsgPack)

	cud.ReactionsCount, err = reactionsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}

	cud.CommentsCount, err = commentsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}

	cud.MeReaction, err = meReaction.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}

	return cud, nil
}

type userSnippetUIData struct {
	User      UITypes.UserSnippet
	MeFollow  bool
	FollowsMe bool
}

func GetUserSnippetUIData(ctx context.Context, username, clientUsername string) (userSnippetUIData, error) {

	var (
		usud                userSnippetUIData
		user                *redis.StringCmd
		meFollow, followsMe *redis.FloatCmd
	)

	_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		user = pipe.HGet(ctx, "users", username)
		meFollow = pipe.ZScore(ctx, fmt.Sprintf("user:%s:followings", clientUsername), username)
		followsMe = pipe.ZScore(ctx, fmt.Sprintf("user:%s:followers", clientUsername), username)

		return nil
	})
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userSnippetUIData{}, err
	}

	userMsgPack, err := user.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userSnippetUIData{}, err
	}

	usud.User = helpers.FromMsgPack[UITypes.UserSnippet](userMsgPack)

	_, err = meFollow.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userSnippetUIData{}, err
	}

	usud.MeFollow = err == nil

	_, err = followsMe.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userSnippetUIData{}, err
	}

	usud.FollowsMe = err == nil

	return usud, nil
}

type reactorSnippetUIData struct {
	User         UITypes.ReactorSnippet
	UserReaction string
}

func GetReactorSnippetUIData(ctx context.Context, username, whichEntity, entityId string) (reactorSnippetUIData, error) {

	var (
		rsud               reactorSnippetUIData
		user, userReaction *redis.StringCmd
	)

	_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		user = pipe.HGet(ctx, "users", username)

		switch whichEntity {
		case "post":
			userReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_post:%s:reactions", entityId), username)
		case "comment":
			userReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", entityId), username)
		}

		return nil
	})
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return reactorSnippetUIData{}, err
	}

	userMsgPack, err := user.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return reactorSnippetUIData{}, err
	}

	rsud.User = helpers.FromMsgPack[UITypes.ReactorSnippet](userMsgPack)

	rsud.UserReaction, err = userReaction.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return reactorSnippetUIData{}, err
	}

	return rsud, nil
}

type userProfileUIData struct {
	User            UITypes.UserProfile
	PostsCount      int64
	FollowersCount  int64
	FollowingsCount int64
	MeFollow        bool
	FollowsMe       bool
}

func GetUserProfileUIData(ctx context.Context, username, clientUsername string) (userProfileUIData, error) {

	var (
		upud                                        userProfileUIData
		user                                        *redis.StringCmd
		postsCount, followersCount, followingsCount *redis.IntCmd
		meFollow, followsMe                         *redis.FloatCmd
	)

	_, err := rdb().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		user = pipe.HGet(ctx, "users", username)
		postsCount = pipe.ZCard(ctx, fmt.Sprintf("user:%s:posts", username))
		followersCount = pipe.ZCard(ctx, fmt.Sprintf("user:%s:followers", username))
		followingsCount = pipe.ZCard(ctx, fmt.Sprintf("user:%s:followings", username))
		meFollow = pipe.ZScore(ctx, fmt.Sprintf("user:%s:followings", clientUsername), username)
		followsMe = pipe.ZScore(ctx, fmt.Sprintf("user:%s:followers", clientUsername), username)

		return nil
	})
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	userMsgPack, err := user.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}
	upud.User = helpers.FromMsgPack[UITypes.UserProfile](userMsgPack)

	upud.PostsCount, err = postsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	upud.FollowersCount, err = followersCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	upud.FollowingsCount, err = followingsCount.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	_, err = meFollow.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	upud.MeFollow = err == nil

	_, err = followsMe.Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	upud.FollowsMe = err == nil

	return upud, nil
}
