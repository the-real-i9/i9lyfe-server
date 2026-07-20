package userDBM

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"

	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/types/UITypes"

	"github.com/gofiber/fiber/v3"
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
	Email         string `msgpack:"email"`
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name" db:"name_"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `msgpack:"bio"`
	Presence      string `msgpack:"presence"`
}

func New(ctx context.Context, email, username, name, bio string, birthday int64, password string) (NewUserT, error) {
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

type LoggedInUserT struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name" db:"name_"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Password      string `msgpack:"-" db:"password_"`
}

func LoginFind(ctx context.Context, uniqueIdent string) (LoggedInUserT, error) {
	user, err := pgDB.QueryRowType[LoggedInUserT](
		ctx,
		/* sql */ `
		SELECT username, name_, profile_pic_url, password_ 
		FROM users 
		WHERE username = $1 OR email = $1
		`, uniqueIdent,
	)
	if err != nil {
		helpers.LogError(err)
		return LoggedInUserT{}, fiber.ErrInternalServerError
	}

	return *user, nil
}

func ChangePassword(ctx context.Context, email string, newPassword string) (string, error) {
	username, err := pgDB.QueryRowField[string](
		ctx,
		/* sql */ `
		UPDATE users
		SET password_ = $2
		WHERE email = $1
		RETURNING username
		`, email, newPassword,
	)
	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	return *username, nil
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

func Follow(ctx context.Context, clientUsername, targetUsername string, at int64) (string, error) {
	followNotifId, err := pgDB.QueryRowField[string](
		ctx,
		/* sql */ `
		SELECT * FROM follow_user ($1, $2, $3)
		`, clientUsername, targetUsername, at,
	)
	if err != nil {
		helpers.LogError(err)
		return "", helpers.HandleDBError(err)
	}

	return *followNotifId, nil
}

func Unfollow(ctx context.Context, clientUsername, targetUsername string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM unfollow_user ($1, $2)
		`, clientUsername, targetUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func GetMentionedPosts(ctx context.Context, clientUsername string, limit int64, cursor int64) ([]*UITypes.Post, error) {
	mentPosts, err := pgDB.QueryRowsType[UITypes.Post](
		ctx,
		/* sql */ `
		SELECT * FROM get_mentioned_posts($1, $2, $3)
		`, clientUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return mentPosts, nil
}

func GetReactedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	reactedPosts, err := pgDB.QueryRowsType[UITypes.Post](
		ctx,
		/* sql */ `
		SELECT * FROM get_reacted_posts($1, $2, $3)
		`, clientUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return reactedPosts, nil
}

func GetSavedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	savedPosts, err := pgDB.QueryRowsType[UITypes.Post](
		ctx,
		/* sql */ `
		SELECT * FROM get_saved_posts($1, $2, $3)
		`, clientUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return savedPosts, nil
}

func GetManyNotifs(ctx context.Context, notifIds []string) ([]*UITypes.NotifSnippet, error) {
	notifs, err := pgDB.QueryRowsType[UITypes.NotifSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM fetch_notifs ($1)
		`, notifIds,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, helpers.HandleDBError(err)
	}

	return notifs, nil
}

func GetMyNotifications(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.NotifSnippet, error) {
	notifs, err := pgDB.QueryRowsType[UITypes.NotifSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_my_notifs ($1, $2, $3)
		`, clientUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, helpers.HandleDBError(err)
	}

	return notifs, nil
}

func ReadNotification(ctx context.Context, clientUsername, notifId string) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		UPDATE notifications
		SET unread = false
		WHERE id_ = $1 AND owner_user = $2
		RETURNING true
		`, notifId, clientUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func GetProfile(ctx context.Context, clientUsername, targetUsername string) (UITypes.UserProfile, error) {
	profile, err := pgDB.QueryRowType[UITypes.UserProfile](
		ctx,
		/* sql */ `
		SELECT * FROM get_user_profile($1, $2)
		`, clientUsername, targetUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return UITypes.UserProfile{}, fiber.ErrInternalServerError
	}

	return *profile, nil
}

func GetFollowers(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.UserSnippet, error) {
	followers, err := pgDB.QueryRowsType[UITypes.UserSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_user_followers($1, $2, $3, $4)
		`, clientUsername, targetUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return followers, nil
}

func GetFollowings(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.UserSnippet, error) {
	followings, err := pgDB.QueryRowsType[UITypes.UserSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_user_followings($1, $2, $3, $4)
		`, clientUsername, targetUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return followings, nil
}

func GetPosts(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	posts, err := pgDB.QueryRowsType[UITypes.Post](
		ctx,
		/* sql */ `
		SELECT * FROM get_user_posts($1, $2, $3, $4)
		`, clientUsername, targetUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return posts, nil
}

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen int64) bool {
	var lastSeenVal any
	if presence == "online" {
		lastSeenVal = nil
	} else {
		lastSeenVal = lastSeen
	}

	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		UPDATE users
		SET presence = $2, last_seen = $3
		WHERE username = $1
		RETURNING true AS done
		`, clientUsername, presence, lastSeenVal,
	)
	if err != nil {
		helpers.LogError(err)
		return false
	}

	return *done
}
