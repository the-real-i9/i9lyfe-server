package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"slices"

	"github.com/redis/go-redis/v9"
)

func GetChat[T any](ctx context.Context, ownerUser, partnerUser string) (chat T, err error) {
	chatJson, err := rdb().HGet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), partnerUser).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return chat, err
	}

	return helpers.FromJson[T](chatJson), nil
}

func GetChatUnreadMsgsCount(ctx context.Context, ownerUser, partnerUser string) (int64, error) {
	count, err := rdb().SCard(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:unread_messages", ownerUser, partnerUser)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return count, nil
}

func GetChatHistoryEntry[T any](ctx context.Context, CHEId string) (CHE T, err error) {
	CHEJson, err := rdb().HGet(ctx, "chat_history_entries", CHEId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return CHE, err
	}

	CHEMap := helpers.FromJson[map[string]any](CHEJson)

	cheType := CHEMap["che_type"].(string)

	if cheType == "message" {
		content := CHEMap["content"].(map[string]any)
		contentProps := content["props"].(map[string]any)

		if content["type"].(string) != "text" {
			mediaCloudName := contentProps["media_cloud_name"].(string)

			if slices.Contains([]string{"photo", "video"}, content["type"].(string)) {
				var (
					blurPlchMcn string
					actualMcn   string
				)

				_, err = fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
				if err != nil {
					return CHE, err
				}

				blurPlchUrl, err := getMediaurl(blurPlchMcn)
				if err != nil {
					return CHE, err
				}

				actualUrl, err := getMediaurl(actualMcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
			} else {
				var mcn string

				_, err = fmt.Sscanf(mediaCloudName, "%s", &mcn)
				if err != nil {
					return CHE, err
				}

				mediaUrl, err := getMediaurl(mcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = mediaUrl
			}

			delete(contentProps, "media_cloud_name")
		}
	}

	return helpers.MapToStruct[T](CHEMap), nil
}

func GetMsgReactions(ctx context.Context, msgId string) (map[string]string, error) {
	msgReactions, err := rdb().HGetAll(ctx, fmt.Sprintf("message:%s:reactions", msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return nil, err
	}

	return msgReactions, nil
}
