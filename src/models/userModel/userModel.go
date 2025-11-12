package userModel

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/models/modelHelpers"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

func Exists(ctx context.Context, uniqueIdent string) (bool, error) {
	userExists, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $1) AS exists
		`, uniqueIdent,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *userExists, nil
}

type NewUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name" db:"name_"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `json:"bio"`
	Presence      string `json:"presence"`
}

func New(ctx context.Context, email, username, password, name, bio string, birthday int64) (NewUserT, error) {
	newUser, err := pgDB.QueryRowType[NewUserT](ctx,
		/* sql */ `
		INSERT INTO users (username, email, password_, name_, bio, birthday)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING email, username, name_, profile_pic_url, bio, presence
		`, username, email, password, name, bio, birthday,
	)
	if err != nil {
		helpers.LogError(err)
		return NewUserT{}, fiber.ErrInternalServerError
	}

	return *newUser, nil
}

type ToAuthUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name" db:"name_"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Presence      string `json:"presence"`
	Password      string `json:"-" db:"password_"`
}

func AuthFind(ctx context.Context, uniqueIdent string) (*ToAuthUserT, error) {
	user, err := pgDB.QueryRowType[ToAuthUserT](
		ctx,
		/* sql */ `
		SELECT email, username, name_, profile_pic_url, presence, password_ 
		FROM users 
		WHERE username = $1 OR email = $1
		`, uniqueIdent,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return user, nil
}

func ChangePassword(ctx context.Context, email, newPassword string) error {
	err := pgDB.Exec(
		ctx,
		/* sql */ `
		UPDATE users
		SET password_ = $2
		WHERE email = $1
		`, email, newPassword,
	)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func EditProfile(ctx context.Context, clientUsername string, updateKVMap map[string]any) (bool, error) {
	if val, ok := updateKVMap["name"]; ok {
		updateKVMap["name_"] = val

		delete(updateKVMap, "name")
	}

	setChanges, params, place := "", []any{clientUsername}, 2

	for col, val := range updateKVMap {
		if setChanges != "" {
			setChanges += ", "
		}
		setChanges += fmt.Sprintf("%s = $%d", col, place)
		params = append(params, val)
		place++
	}

	done, err := pgDB.QueryRowField[bool](
		ctx,
		fmt.Sprintf( /* sql */ `
		UPDATE users
		SET %s 
		WHERE username = $1
		RETURNING true AS done
		`, setChanges), params...,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func ChangeProfilePicture(ctx context.Context, clientUsername, pictureUrl string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		UPDATE users
		SET profile_pic_url = $2
		WHERE username = $1
		RETURNING true AS done
		`, clientUsername, pictureUrl,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func Follow(ctx context.Context, clientUsername, targetUsername string, at int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		INSERT INTO user_follows_user (follower_username, following_username, at_)
		VALUES ($1, $2, $3)
		RETURNING true AS done
		`, clientUsername, targetUsername, at,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func Unfollow(ctx context.Context, clientUsername, targetUsername string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_follows_user
		WHERE follower_username = $1 AND following_username = $2
		RETURNING true AS done
		`, clientUsername, targetUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func GetMentionedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.Post, error) {
	mentPostsMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:mentioned_posts", clientUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	mentPosts, err := modelHelpers.PostMembersForUIPosts(ctx, mentPostsMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return mentPosts, nil
}

func GetReactedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.Post, error) {
	reactedPostsMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:reacted_posts", clientUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	reactedPosts, err := modelHelpers.PostMembersForUIPosts(ctx, reactedPostsMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return reactedPosts, nil
}

func GetSavedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.Post, error) {
	savedPostsMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:saved_posts", clientUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	savedPosts, err := modelHelpers.PostMembersForUIPosts(ctx, savedPostsMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return savedPosts, nil
}

func GetNotifications(ctx context.Context, clientUsername string, year int, month string, limit int, cursor float64) ([]UITypes.NotifSnippet, error) {
	notifsMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:notifications:%d-%s", clientUsername, year, month), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	notifs, err := modelHelpers.NotifMembersForUINotifSnippets(ctx, notifsMembers)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return notifs, nil
}

func ReadNotification(ctx context.Context, clientUsername, year, month, notifId string) (bool, error) {
	itIs, err := cache.IsUserNotification(ctx, clientUsername, year, month, notifId)
	if err != nil {
		return false, fiber.ErrInternalServerError
	}

	if !itIs {
		return false, nil
	}

	err = cache.RemoveUnreadNotifications(ctx, notifId)
	if err != nil {
		return false, fiber.ErrInternalServerError
	}

	return true, nil
}

func GetProfile(ctx context.Context, clientUsername, targetUsername string) (UITypes.UserProfile, error) {
	userProfile, err := modelHelpers.BuildUserProfileUIFromCache(ctx, targetUsername, clientUsername)
	if err != nil {
		return UITypes.UserProfile{}, err
	}

	return userProfile, nil
}

func GetFollowers(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.UserSnippet, error) {
	followerMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:followers", targetUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	followers, err := modelHelpers.UserMembersForUIUserSnippets(ctx, followerMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return followers, nil
}

func GetFollowings(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.UserSnippet, error) {
	followingMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:followings", targetUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	followings, err := modelHelpers.UserMembersForUIUserSnippets(ctx, followingMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return followings, nil
}

func GetPosts(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.Post, error) {
	userPostsMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:posts", targetUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	userPosts, err := modelHelpers.PostMembersForUIPosts(ctx, userPostsMembers, clientUsername)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return userPosts, nil
}

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen int64) error {
	var lastSeenVal any
	if presence == "online" {
		lastSeenVal = nil
	} else {
		lastSeenVal = lastSeen
	}

	err := pgDB.Exec(
		ctx,
		/* sql */ `
		UPDATE users
		SET presence = $2, last_seen = $3
		WHERE username = $1
		`, clientUsername, presence, lastSeenVal,
	)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
