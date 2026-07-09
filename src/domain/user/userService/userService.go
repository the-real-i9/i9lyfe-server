package userService

import (
	"context"
	"fmt"

	"i9lyfe/src/appGlobals"
	user "i9lyfe/src/domain/user/userDBM"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/mediaStorageService"
	"i9lyfe/src/services/pubsubService"
	"i9lyfe/src/services/sseService"
	"i9lyfe/src/types/UITypes"
	"i9lyfe/src/types/appTypes"
	"i9lyfe/src/types/eventTypes"
	"maps"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

func dbPool() *pgxpool.Pool {
	return appGlobals.DBPool
}

func UserExists(ctx context.Context, uniqueIdent string) (bool, error) {
	return user.Exists(ctx, uniqueIdent)
}

func NewUser(ctx context.Context, email, username, name, bio string, birthday int64, password string) (user.NewUserT, error) {
	newUser, err := user.New(ctx, email, username, name, bio, birthday, password)
	if err != nil {
		return newUser, err
	}

	newUser.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(newUser.ProfilePicUrl)

	return newUser, nil
}

func LoginUserFind(ctx context.Context, uniqueIdent string) (user.LoggedInUserT, error) {
	fUser, err := user.LoginFind(ctx, uniqueIdent)
	if err != nil {
		return fUser, err
	}

	if fUser.Username != "" {
		fUser.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(fUser.ProfilePicUrl)
	}

	return fUser, nil
}

func ChangeUserPassword(ctx context.Context, email string, newPassword string) (bool, error) {
	username, err := user.ChangePassword(ctx, email, newPassword)
	if err != nil {
		return false, err
	}

	done := username != ""

	return done, nil
}

func EditUserProfile(ctx context.Context, clientUsername string, updateKVStruct any) (bool, error) {
	updateKVMap := helpers.StructToMap(updateKVStruct)

	done, err := user.EditProfile(ctx, clientUsername, maps.Clone(updateKVMap))
	if err != nil {
		return false, err
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

		pPicCloudName := fmt.Sprintf("uploads/user/profile_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), utils.UUIDv4(), which[small0_medium1_large2])

		url, err := mediaStorageService.GetUploadUrl(pPicCloudName, picMIME)
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

	return done, nil
}

func FollowUser(ctx context.Context, clientUsername, targetUsername string, at int64) (bool, error) {
	if clientUsername == targetUsername {
		return false, fiber.NewError(fiber.StatusBadRequest, "are you trying to follow yourself???")
	}

	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	followNotifId, err := user.Follow(ctx, pgTx, clientUsername, targetUsername, at)
	if err != nil {
		return false, err
	}

	done := followNotifId != ""

	if done {
		err = eventStreamService.QueueUserFollowEvent(ctx, eventTypes.UserFollowEvent{
			FollowerUser:  clientUsername,
			FollowingUser: targetUsername,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}

		err = pgTx.Commit(ctx)
		if err != nil {
			helpers.LogError(err)
			return false, fiber.ErrInternalServerError
		}

		go func() {
			notif, err := GetOneNotif(ctx, followNotifId)
			if err != nil {
				return
			}

			sseService.SendEventMsg(targetUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  notif,
			})
		}()
	}

	return done, nil
}

func UnfollowUser(ctx context.Context, clientUsername, targetUsername string) (any, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := user.Unfollow(ctx, pgTx, clientUsername, targetUsername)
	if err != nil {
		return nil, err
	}

	if done {
		err = eventStreamService.QueueUserUnfollowEvent(ctx, eventTypes.UserUnfollowEvent{
			FollowerUser:  clientUsername,
			FollowingUser: targetUsername,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func GetUserMentionedPosts(ctx context.Context, clientUsername string, limit int64, cursor int64) ([]*UITypes.Post, error) {
	posts, err := user.GetMentionedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, p := range posts {
		p.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(p.OwnerUser["profile_pic_url"].(string))

		p.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(p.MediaUrls)
	}

	return posts, nil
}

func GetUserReactedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	posts, err := user.GetReactedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, p := range posts {
		p.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(p.OwnerUser["profile_pic_url"].(string))

		p.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(p.MediaUrls)
	}

	return posts, nil
}

func GetUserSavedPosts(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	posts, err := user.GetSavedPosts(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, p := range posts {
		p.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(p.OwnerUser["profile_pic_url"].(string))

		p.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(p.MediaUrls)
	}

	return posts, nil
}

func GetManyNotifs(ctx context.Context, notifIds []string) ([]*UITypes.NotifSnippet, error) {
	notifs, err := user.GetManyNotifs(ctx, notifIds)
	if err != nil {
		return nil, err
	}

	mediaStorageService.NotifMediaCloudNamesToUrl(notifs)

	return notifs, nil
}
func GetOneNotif(ctx context.Context, notifId string) (UITypes.NotifSnippet, error) {
	n, err := GetManyNotifs(ctx, []string{notifId})
	if err != nil {
		return UITypes.NotifSnippet{}, err
	}

	return *(n[0]), nil
}

func GetUserNotifications(ctx context.Context, clientUsername string, limit int64, cursor float64) ([]*UITypes.NotifSnippet, error) {
	notifs, err := user.GetMyNotifications(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	mediaStorageService.NotifMediaCloudNamesToUrl(notifs)

	return notifs, nil
}

func ReadUserNotification(ctx context.Context, clientUsername, notifId string) (bool, error) {
	return user.ReadNotification(ctx, clientUsername, notifId)

}

func GetUserProfile(ctx context.Context, clientUsername, targetUsername string) (UITypes.UserProfile, error) {
	p, err := user.GetProfile(ctx, clientUsername, targetUsername)
	if err != nil {
		return UITypes.UserProfile{}, err
	}

	p.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(p.ProfilePicUrl)

	return p, nil
}

func GetUserFollowers(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.UserSnippet, error) {
	folls, err := user.GetFollowers(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, f := range folls {
		f.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(f.ProfilePicUrl)
	}

	return folls, nil
}

func GetUserFollowings(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.UserSnippet, error) {
	folls, err := user.GetFollowings(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, f := range folls {
		f.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(f.ProfilePicUrl)
	}

	return folls, nil
}

func GetUserPosts(ctx context.Context, clientUsername, targetUsername string, limit int64, cursor float64) ([]*UITypes.Post, error) {
	posts, err := user.GetPosts(ctx, clientUsername, targetUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, p := range posts {
		p.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(p.OwnerUser["profile_pic_url"].(string))

		p.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(p.MediaUrls)
	}

	return posts, nil
}

func GoOnline(ctx context.Context, clientUsername string) {
	done := user.ChangePresence(ctx, clientUsername, "online", 0)

	if done {
		pubsubService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":     clientUsername,
			"presence": "online",
		})
	}
}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC().UnixMilli()

	done := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)

	if done {
		pubsubService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":      clientUsername,
			"presence":  "offline",
			"last_seen": lastSeen,
		})
	}

}
