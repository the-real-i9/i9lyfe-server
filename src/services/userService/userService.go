package userService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	user "i9lyfe/src/models/userModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func EditUserProfile(ctx context.Context, clientUsername string, updateKVStruct any) (bool, error) {
	updateKVMap := helpers.StructToMap(updateKVStruct)

	done, err := user.EditProfile(ctx, clientUsername, updateKVMap)
	if err != nil {
		return false, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: updateKVMap,
		})
	}

	return done, nil
}

func ChangeUserProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, error) {

	mime := mimetype.Detect(pictureData)
	fileType := mime.String()
	fileExt := mime.Extension()

	if !strings.HasPrefix(fileType, "image") {
		return nil, fiber.NewError(400, fmt.Sprintf("invalid file type %s, for picture_data, expected image/*", fileType))
	}

	pictureUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("profile_pictures/%s/ppic-%d%s", clientUsername, time.Now().UnixNano(), fileExt), pictureData)
	if err != nil {
		return nil, err
	}

	done, err := user.ChangeProfilePicture(ctx, clientUsername, pictureUrl)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_url": pictureUrl},
		})
	}

	return done, nil
}

func FollowUser(ctx context.Context, clientUsername, targetUsername string, at int64) (any, error) {
	if clientUsername == targetUsername {
		return nil, fiber.NewError(fiber.StatusBadRequest, "are you trying to follow yourself???")
	}

	done, err := user.Follow(ctx, clientUsername, targetUsername, at)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueUserFollowEvent(eventTypes.UserFollowEvent{
			FollowerUser:  clientUsername,
			FollowingUser: targetUsername,
			At:            at,
		})
	}

	return done, nil
}

func UnfollowUser(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	done, err := user.Unfollow(ctx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueUserUnfollowEvent(eventTypes.UserUnfollowEvent{
			FollowerUser:  clientUsername,
			FollowingUser: targetUsername,
		})
	}

	return done, nil
}

func GetUserMentionedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) (any, error) {
	posts, err := user.GetMentionedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserReactedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) (any, error) {
	posts, err := user.GetReactedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserSavedPosts(ctx context.Context, clientUsername string, limit int, cursor float64) (any, error) {
	posts, err := user.GetSavedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserNotifications(ctx context.Context, clientUsername string, year int, month string, limit int, cursor float64) (any, error) {
	notifs, err := user.GetNotifications(ctx, clientUsername, year, month, limit, cursor)
	if err != nil {
		return nil, err
	}

	return notifs, nil
}

func ReadUserNotification(ctx context.Context, clientUsername, year, month, notifId string) (bool, error) {
	done, err := user.ReadNotification(ctx, clientUsername, year, month, notifId)
	if err != nil {
		return false, err
	}

	return done, nil
}

func GetUserProfile(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	profile, err := user.GetProfile(ctx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func GetUserFollowers(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.UserSnippet, error) {
	followers, err := user.GetFollowers(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return followers, nil
}

func GetUserFollowings(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.UserSnippet, error) {
	followings, err := user.GetFollowings(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return followings, nil
}

func GetUserPosts(ctx context.Context, clientUsername, targetUsername string, limit int, cursor float64) ([]UITypes.Post, error) {
	posts, err := user.GetPosts(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GoOnline(ctx context.Context, clientUsername string) {
	err := user.ChangePresence(ctx, clientUsername, "online", time.Time{})
	if err != nil {
		return
	}

	realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
		"user":     clientUsername,
		"presence": "online",
	})
}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC()

	err := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)
	if err != nil {
		return
	}

	realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
		"user":      clientUsername,
		"presence":  "offline",
		"last_seen": lastSeen,
	})
}
