package userService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	user "i9lyfe/src/models/userModel"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"maps"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthPPicDataT struct {
	UploadUrl     string `json:"uploadUrl"`
	PPicCloudName string `json:"profilePicCloudName"`
}

func AuthorizePPicUpload(ctx context.Context, picMIME string, picSize [3]int64) (AuthPPicDataT, error) {
	var res AuthPPicDataT

	for small0_medium1_large2, size := range picSize {

		which := [3]string{"small", "medium", "large"}

		pPicCloudName := fmt.Sprintf("uploads/user/profile_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), which[small0_medium1_large2])

		url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
			pPicCloudName,
			&storage.SignedURLOptions{
				Scheme:      storage.SigningSchemeV4,
				Method:      "PUT",
				ContentType: picMIME,
				Expires:     time.Now().Add(15 * time.Minute),
				Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", size)},
			},
		)
		if err != nil {
			helpers.LogError(err)
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

func ChangeUserProfilePicture(ctx context.Context, clientUsername, profilePicCloudName string) (any, error) {
	done, err := user.ChangeProfilePicture(ctx, clientUsername, profilePicCloudName)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_cloud_name": profilePicCloudName},
		})
	}

	return done, nil
}

func FollowUser(ctx context.Context, clientUser appTypes.ClientUser, targetUsername string, at int64) (any, error) {
	if clientUser.Username == targetUsername {
		return nil, fiber.NewError(fiber.StatusBadRequest, "are you trying to follow yourself???")
	}

	done, err := user.Follow(ctx, clientUser.Username, targetUsername, at)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueUserFollowEvent(eventTypes.UserFollowEvent{
			FollowerUser:  clientUser,
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
	done := user.ChangePresence(ctx, clientUsername, "online", 0)

	if done {
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
		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":      clientUsername,
			"presence":  "offline",
			"last_seen": lastSeen,
		})
	}

}
