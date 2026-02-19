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
	"maps"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func UserExists(ctx context.Context, uniqueIdent string) (bool, error) {
	return user.Exists(ctx, uniqueIdent)
}

func NewUser(ctx context.Context, email, username, name, bio string, birthday int64, password string) (user.NewUserT, error) {
	newUser, err := user.New(ctx, email, username, name, bio, birthday, password)
	if err != nil {
		return newUser, err
	}

	go eventStreamService.QueueNewUserEvent(eventTypes.NewUserEvent{
		Username: newUser.Username,
		UserData: helpers.ToMsgPack(newUser),
	})

	newUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(newUser.ProfilePicUrl)

	return newUser, nil
}

func SigninUserFind(ctx context.Context, uniqueIdent string) (user.SignedInUserT, error) {
	fUser, err := user.SigninFind(ctx, uniqueIdent)
	if err != nil {
		return fUser, err
	}

	if fUser.Username != "" {
		fUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(fUser.ProfilePicUrl)
	}

	return fUser, nil
}

func ChangeUserPassword(ctx context.Context, email string, newPassword string) (bool, error) {
	username, err := user.ChangePassword(ctx, email, newPassword)
	if err != nil {
		return false, err
	}

	done := username != ""

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    username,
			UpdateKVMap: map[string]any{"password": newPassword},
		})
	}

	return done, nil
}

func EditUserProfile(ctx context.Context, clientUsername string, updateKVStruct any) (bool, error) {
	updateKVMap := helpers.StructToMap(updateKVStruct)

	done, err := user.EditProfile(ctx, clientUsername, maps.Clone(updateKVMap))
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

type AuthPPicDataT struct {
	UploadUrl     string `msgpack:"uploadUrl"`
	PPicCloudName string `msgpack:"profilePicCloudName"`
}

func AuthorizePPicUpload(ctx context.Context, picMIME string) (AuthPPicDataT, error) {
	var res AuthPPicDataT

	for small0_medium1_large2 := range 3 {

		which := [3]string{"small", "medium", "large"}

		pPicCloudName := fmt.Sprintf("uploads/user/profile_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), which[small0_medium1_large2])

		url, err := cloudStorageService.GetUploadUrl(pPicCloudName, picMIME)
		if err != nil {
			return AuthPPicDataT{}, fiber.ErrInternalServerError
		}

		switch small0_medium1_large2 {
		case 0:
			res.UploadUrl += "small:"
			res.PPicCloudName += "small:"
		case 1:
			res.UploadUrl += "medium:"
			res.PPicCloudName += "medium:"
		default:
			res.UploadUrl += "large:"
			res.PPicCloudName += "large:"
		}

		res.UploadUrl += url
		res.PPicCloudName += pPicCloudName

		if small0_medium1_large2 != 2 {
			res.UploadUrl += " "
			res.PPicCloudName += " "
		}
	}

	return res, nil
}

func ChangeUserProfilePicture(ctx context.Context, clientUsername, profilePicCloudName string) (bool, error) {
	done, err := user.ChangeProfilePicture(ctx, clientUsername, profilePicCloudName)
	if err != nil {
		return false, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_url": profilePicCloudName},
		})
	}

	return done, nil
}

func FollowUser(ctx context.Context, clientUsername, targetUsername string, at int64) (any, error) {
	if clientUsername == targetUsername {
		return nil, fiber.NewError(fiber.StatusBadRequest, "are you trying to follow yourself???")
	}

	followCursor, err := user.Follow(ctx, clientUsername, targetUsername, at)
	if err != nil {
		return nil, err
	}

	done := followCursor != 0

	if done {
		go eventStreamService.QueueUserFollowEvent(eventTypes.UserFollowEvent{
			FollowerUser:  clientUsername,
			FollowingUser: targetUsername,
			At:            at,
			FollowCursor:  followCursor,
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

func GetUserMentionedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) (any, error) {
	return user.GetMentionedPosts(ctx, clientUsername, limit, cursor)

}

func GetUserReactedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) (any, error) {
	return user.GetReactedPosts(ctx, clientUsername, limit, cursor)
}

func GetUserSavedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) (any, error) {
	return user.GetSavedPosts(ctx, clientUsername, limit, cursor)

}

func GetUserNotifications(ctx context.Context, clientUsername string, year int64, month int64, limit int64, cursor float64) (any, error) {
	return user.GetNotifications(ctx, clientUsername, year, month, limit, cursor)

}

func ReadUserNotification(ctx context.Context, clientUsername, year, month, notifId string) (bool, error) {
	return user.ReadNotification(ctx, clientUsername, year, month, notifId)

}

func GetUserProfile(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	return user.GetProfile(ctx, clientUsername, targetUsername)

}

func GetUserFollowers(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]UITypes.UserSnippet, error) {
	return user.GetFollowers(ctx, clientUsername, targetUsername, limit, cursor)
}

func GetUserFollowings(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]UITypes.UserSnippet, error) {
	return user.GetFollowings(ctx, clientUsername, targetUsername, limit, cursor)
}

func GetUserPosts(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]UITypes.Post, error) {
	return user.GetPosts(ctx, clientUsername, targetUsername, limit, cursor)
}

func GoOnline(ctx context.Context, clientUsername string) {
	done := user.ChangePresence(ctx, clientUsername, "online", 0)

	if done {
		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "online",
			LastSeen: 0,
		})

		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":     clientUsername,
			"presence": "online",
		})
	}
}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC().UnixMilli()

	done := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)

	if done {
		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "offline",
			LastSeen: lastSeen,
		})

		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":      clientUsername,
			"presence":  "offline",
			"last_seen": lastSeen,
		})
	}

}
