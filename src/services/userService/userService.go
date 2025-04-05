package userService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	user "i9lyfe/src/models/userModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/messageBrokerService"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func GetClientUser(ctx context.Context, clientUsername string) (any, error) {
	clientUser, err := user.Client(ctx, clientUsername)
	if err != nil {
		return nil, err
	}

	return clientUser, nil
}

func EditUserProfile(ctx context.Context, clientUsername string, updateKVStruct any) (any, error) {
	var updateKVMap map[string]any

	helpers.AnyToAny(updateKVStruct, &updateKVMap)

	err := user.EditProfile(ctx, clientUsername, updateKVMap)
	if err != nil {
		return nil, err
	}

	return appGlobals.OprSucc, nil
}

func ChangeUserProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, error) {

	mime := mimetype.Detect(pictureData)
	fileType := mime.String()
	fileExt := mime.Extension()

	if !strings.HasPrefix(fileType, "image") {
		return nil, fiber.NewError(400, fmt.Sprintf("invalid file type %s, for picture_data, expected image/*", fileType))
	}

	pictureUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("profile_pictures/%s", clientUsername), pictureData, fileExt)
	if err != nil {
		return nil, err
	}

	err = user.ChangeProfilePicture(ctx, clientUsername, pictureUrl)
	if err != nil {
		return nil, err
	}

	return appGlobals.OprSucc, nil
}

func FollowUser(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	if clientUsername == targetUsername {
		return nil, fiber.NewError(fiber.StatusBadRequest, "are you trying to follow yourself???")
	}

	followNotif, err := user.Follow(ctx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	go func(followNotif map[string]any) {
		if fn := followNotif; fn != nil {
			receiverUsername := fn["receiver_username"].(string)

			delete(fn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  fn,
			})
		}
	}(followNotif)

	return appGlobals.OprSucc, nil
}

func UnfollowUser(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	err := user.Unfollow(ctx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	return appGlobals.OprSucc, nil
}

func GetUserMentionedPosts(ctx context.Context, clientUsername string, limit int, offset int64) (any, error) {
	posts, err := user.GetMentionedPosts(ctx, clientUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserReactedPosts(ctx context.Context, clientUsername string, limit int, offset int64) (any, error) {
	posts, err := user.GetReactedPosts(ctx, clientUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserSavedPosts(ctx context.Context, clientUsername string, limit int, offset int64) (any, error) {
	posts, err := user.GetSavedPosts(ctx, clientUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserNotifications(ctx context.Context, clientUsername string, limit int, offset int64) (any, error) {
	notifs, err := user.GetNotifications(ctx, clientUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return notifs, nil
}

func ReadUserNotification(ctx context.Context, clientUsername, notificationId string) (any, error) {
	err := user.ReadNotification(ctx, clientUsername, notificationId)
	if err != nil {
		return nil, err
	}

	return appGlobals.OprSucc, nil
}

func GetUserProfile(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	profile, err := user.GetProfile(ctx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
func GetUserFollowers(ctx context.Context, clientUsername, targetUsername string, limit int, offset int64) (any, error) {
	profile, err := user.GetFollowers(ctx, clientUsername, targetUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return profile, nil
}
func GetUserFollowing(ctx context.Context, clientUsername, targetUsername string, limit int, offset int64) (any, error) {
	profile, err := user.GetFollowing(ctx, clientUsername, targetUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return profile, nil
}
func GetUserPosts(ctx context.Context, clientUsername, targetUsername string, limit int, offset int64) (any, error) {
	profile, err := user.GetPosts(ctx, clientUsername, targetUsername, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return profile, nil
}
