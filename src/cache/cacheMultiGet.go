package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

type postUIData struct {
	OwnerUser      func() (UITypes.ContentOwnerUser, error)
	ReposterUser   func() (UITypes.ClientUser, error)
	ReactionsCount func() (int64, error)
	CommentsCount  func() (int64, error)
	RepostsCount   func() (int64, error)
	SavesCount     func() (int64, error)
	MeReaction     func() (string, error)
	MeSaved        func() (bool, error)
	MeReposted     func() (bool, error)
}

func GetPostUIData(ctx context.Context, postId, ownerUsername, reposterUsername, clientUsername string) (postUIData, error) {
	var pud postUIData

	pipe := rdb().Pipeline()

	ownerUser := pipe.HGet(ctx, "users", ownerUsername)
	reposterUser := pipe.HGet(ctx, "users", reposterUsername)
	reactionsCount := pipe.HLen(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId))
	commentsCount := pipe.ZCard(ctx, fmt.Sprintf("commented_post:%s:comments", postId))
	repostsCount := pipe.SCard(ctx, fmt.Sprintf("reposted_post:%s:reposts", postId))
	savesCount := pipe.SCard(ctx, fmt.Sprintf("saved_post:%s:saves", postId))
	meReaction := pipe.HGet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), clientUsername)
	meReposted := pipe.ZScore(ctx, fmt.Sprintf("user:%s:reposted_posts", clientUsername), postId)
	meSaved := pipe.ZScore(ctx, fmt.Sprintf("user:%s:saved_posts", clientUsername), postId)

	_, err := pipe.Exec(ctx)
	if err != nil {
		helpers.LogError(err)
		return postUIData{}, err
	}

	userJson, err := ownerUser.Result()
	pud.OwnerUser = func() (UITypes.ContentOwnerUser, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.ContentOwnerUser{}, err
		}

		return helpers.FromJson[UITypes.ContentOwnerUser](userJson), nil
	}

	userJson, err = reposterUser.Result()
	pud.ReposterUser = func() (UITypes.ClientUser, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.ClientUser{}, err
		}

		return helpers.FromJson[UITypes.ClientUser](userJson), nil
	}

	count, err := reactionsCount.Result()
	pud.ReactionsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = commentsCount.Result()
	pud.CommentsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = repostsCount.Result()
	pud.RepostsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = savesCount.Result()
	pud.SavesCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	rxn, err := meReaction.Result()
	pud.MeReaction = func() (string, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return rxn, err
		}

		return rxn, nil
	}

	_, err = meReposted.Result()
	pud.MeReposted = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	_, err = meSaved.Result()
	pud.MeSaved = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	return pud, nil
}

type commentUIData struct {
	OwnerUser      func() (UITypes.ContentOwnerUser, error)
	ReactionsCount func() (int64, error)
	CommentsCount  func() (int64, error)
	MeReaction     func() (string, error)
}

func GetCommentUIData(ctx context.Context, commentId, ownerUsername, clientUsername string) (commentUIData, error) {
	var cud commentUIData

	pipe := rdb().Pipeline()

	ownerUser := pipe.HGet(ctx, "users", ownerUsername)
	reactionsCount := pipe.HLen(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId))
	commentsCount := pipe.ZCard(ctx, fmt.Sprintf("commented_comment:%s:comments", commentId))
	meReaction := pipe.HGet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), clientUsername)

	_, err := pipe.Exec(ctx)
	if err != nil {
		helpers.LogError(err)
		return commentUIData{}, err
	}

	userJson, err := ownerUser.Result()
	cud.OwnerUser = func() (UITypes.ContentOwnerUser, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.ContentOwnerUser{}, err
		}

		return helpers.FromJson[UITypes.ContentOwnerUser](userJson), nil
	}

	count, err := reactionsCount.Result()
	cud.ReactionsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = commentsCount.Result()
	cud.CommentsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	rxn, err := meReaction.Result()
	cud.MeReaction = func() (string, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return rxn, err
		}

		return rxn, nil
	}

	return cud, nil
}

type userSnippetUIData struct {
	User      func() (UITypes.UserSnippet, error)
	MeFollow  func() (bool, error)
	FollowsMe func() (bool, error)
}

func GetUserSnippetUIData(ctx context.Context, username, clientUsername string) (userSnippetUIData, error) {
	var usud userSnippetUIData

	pipe := rdb().Pipeline()

	user := pipe.HGet(ctx, "users", username)
	meFollow := pipe.ZScore(ctx, fmt.Sprintf("user:%s:followings", clientUsername), username)
	followsMe := pipe.ZScore(ctx, fmt.Sprintf("user:%s:followers", clientUsername), username)

	_, err := pipe.Exec(ctx)
	if err != nil {
		helpers.LogError(err)
		return userSnippetUIData{}, err
	}

	userJson, err := user.Result()
	usud.User = func() (UITypes.UserSnippet, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.UserSnippet{}, err
		}

		return helpers.FromJson[UITypes.UserSnippet](userJson), nil
	}

	_, err = meFollow.Result()
	usud.MeFollow = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	_, err = followsMe.Result()
	usud.FollowsMe = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	return usud, nil
}

type reactorSnippetUIData struct {
	User         func() (UITypes.ReactorSnippet, error)
	UserReaction func() (string, error)
}

func GetReactorSnippetUIData(ctx context.Context, username, whichEntity, entityId string) (reactorSnippetUIData, error) {
	var rsud reactorSnippetUIData

	pipe := rdb().Pipeline()

	user := pipe.HGet(ctx, "users", username)
	var userReaction *redis.StringCmd

	switch whichEntity {
	case "post":
		userReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_post:%s:reactions", entityId), username)
	case "comment":
		userReaction = pipe.HGet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", entityId), username)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		helpers.LogError(err)
		return reactorSnippetUIData{}, err
	}

	userJson, err := user.Result()
	rsud.User = func() (UITypes.ReactorSnippet, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.ReactorSnippet{}, err
		}

		return helpers.FromJson[UITypes.ReactorSnippet](userJson), nil
	}

	rxn, err := userReaction.Result()
	rsud.UserReaction = func() (string, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return rxn, err
		}

		return rxn, nil
	}

	return rsud, nil
}

type userProfileUIData struct {
	User            func() (UITypes.UserProfile, error)
	PostsCount      func() (int64, error)
	FollowersCount  func() (int64, error)
	FollowingsCount func() (int64, error)
	MeFollow        func() (bool, error)
	FollowsMe       func() (bool, error)
}

func GetUserProfileUIData(ctx context.Context, username, clientUsername string) (userProfileUIData, error) {
	var upud userProfileUIData

	pipe := rdb().Pipeline()

	user := pipe.HGet(ctx, "users", username)
	postsCount := pipe.ZCard(ctx, fmt.Sprintf("user:%s:posts", username))
	followersCount := pipe.ZCard(ctx, fmt.Sprintf("user:%s:followers", username))
	followingsCount := pipe.ZCard(ctx, fmt.Sprintf("user:%s:followings", username))
	meFollow := pipe.ZScore(ctx, fmt.Sprintf("user:%s:followings", clientUsername), username)
	followsMe := pipe.ZScore(ctx, fmt.Sprintf("user:%s:followers", clientUsername), username)

	_, err := pipe.Exec(ctx)
	if err != nil {
		helpers.LogError(err)
		return userProfileUIData{}, err
	}

	userJson, err := user.Result()
	upud.User = func() (UITypes.UserProfile, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return UITypes.UserProfile{}, err
		}

		return helpers.FromJson[UITypes.UserProfile](userJson), nil
	}

	count, err := postsCount.Result()
	upud.PostsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = followersCount.Result()
	upud.FollowersCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	count, err = followingsCount.Result()
	upud.FollowingsCount = func() (int64, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return count, err
		}

		return count, nil
	}

	_, err = meFollow.Result()
	upud.MeFollow = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	_, err = followsMe.Result()
	upud.FollowsMe = func() (bool, error) {
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return false, err
		}

		return err == nil, nil
	}

	return upud, nil
}
