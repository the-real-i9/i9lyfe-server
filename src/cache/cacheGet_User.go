package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/redis/go-redis/v9"
)

func getMediaurl(mcn string) (string, error) {
	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(mcn, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add((6 * 24) * time.Hour),
	})
	if err != nil {
		return "", err
	}

	return url, nil
}

func GetUser[T any](ctx context.Context, username string) (user T, err error) {
	userJson, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return user, err
	}

	userMap := helpers.FromJson[map[string]any](userJson)

	ppicCloudName := userMap["profile_pic_cloud_name"].(string)

	var (
		smallPPicn  string
		mediumPPicn string
		largePPicn  string
	)

	_, err = fmt.Sscanf(ppicCloudName, "small:%s medium:%s large:%s", &smallPPicn, &mediumPPicn, &largePPicn)
	if err != nil {
		return user, err
	}

	smallPicUrl, err := getMediaurl(smallPPicn)
	if err != nil {
		return user, err
	}

	mediumPicUrl, err := getMediaurl(mediumPPicn)
	if err != nil {
		return user, err
	}

	largePicUrl, err := getMediaurl(largePPicn)
	if err != nil {
		return user, err
	}

	userMap["profile_pic_url"] = fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)

	delete(userMap, "profile_pic_cloud_name")

	return helpers.MapToStruct[T](userMap), nil
}

func GetUserPostReaction(ctx context.Context, postId, username string) (string, error) {
	reaction, err := rdb().HGet(ctx, fmt.Sprintf("reacted_post:%s:reactions", postId), username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return "", err
	}

	return reaction, nil
}

func GetUserCommentReaction(ctx context.Context, commentId, username string) (string, error) {
	reaction, err := rdb().HGet(ctx, fmt.Sprintf("reacted_comment:%s:reactions", commentId), username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return "", err
	}

	return reaction, nil
}

func UserSavedPost(ctx context.Context, username, postId string) (bool, error) {
	_, err := rdb().ZScore(ctx, fmt.Sprintf("user:%s:saved_posts", username), postId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return err == nil, nil
}

func UserRepostedPost(ctx context.Context, username, postId string) (bool, error) {
	_, err := rdb().ZScore(ctx, fmt.Sprintf("user:%s:reposted_posts", username), postId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return err == nil, nil
}

func MeFollowUser(ctx context.Context, me, username string) (bool, error) {
	_, err := rdb().ZScore(ctx, fmt.Sprintf("user:%s:followings", me), username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return err == nil, nil
}

func UserFollowsMe(ctx context.Context, me, username string) (bool, error) {
	_, err := rdb().ZScore(ctx, fmt.Sprintf("user:%s:followers", me), username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return err == nil, nil
}

func GetUserPostsCount(ctx context.Context, username string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("user:%s:posts", username)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetUserFollowersCount(ctx context.Context, username string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("user:%s:followers", username)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetUserFollowingsCount(ctx context.Context, username string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("user:%s:followings", username)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func IsUserNotification(ctx context.Context, username, year, month, notifId string) (bool, error) {
	_, err := rdb().ZScore(ctx, fmt.Sprintf("user:%s:notifications:%s-%s", username, year, month), notifId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return err == nil, nil
}

func GetNotification[T any](ctx context.Context, notifId string) (notif T, err error) {
	notifJson, err := rdb().HGet(ctx, "notifications", notifId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return notif, err
	}

	return helpers.FromJson[T](notifJson), nil
}

func NotificationIsUnread(ctx context.Context, notifId string) (bool, error) {
	check, err := rdb().SIsMember(ctx, "unread_notifications", notifId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return false, err
	}

	return check, nil
}
