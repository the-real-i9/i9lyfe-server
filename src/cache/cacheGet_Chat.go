package cache

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"

	"github.com/redis/go-redis/v9"
)

func GetChat[T any](ctx context.Context, ownerUser, partnerUser string) (chat T, err error) {
	chatJson, err := rdb().HGet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), partnerUser).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return chat, err
	}

	if err == redis.Nil {
		dbChat, err := GetChatFromDB[T](ctx, ownerUser, partnerUser)
		if err != nil {
			return chat, err
		}

		return *dbChat, nil
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

	if err == redis.Nil {
		dbCHE, err := GetChatHistoryEntryFromDB[T](ctx, CHEId)
		if err != nil {
			return CHE, err
		}

		return *dbCHE, nil
	}

	return helpers.FromJson[T](CHEJson), nil
}

func GetMsgReactions(ctx context.Context, msgId string) (map[string]string, error) {
	msgReactions, err := rdb().HGetAll(ctx, fmt.Sprintf("message:%s:reactions", msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return msgReactions, err
	}

	return msgReactions, nil
}
