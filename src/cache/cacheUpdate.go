package cache

import (
	"context"
	"i9lyfe/src/helpers"
	"maps"
)

func UpdateUser(ctx context.Context, user string, updateKVMap map[string]any) error {
	userDataJson, err := rdb().HGet(ctx, "users", user).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	userData := helpers.FromJson[map[string]any](userDataJson)

	maps.Copy(userData, updateKVMap)

	err = rdb().HSet(ctx, "users", []string{user, helpers.ToJson(userData)}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateMessage(ctx context.Context, CHEId string, updateKVMap map[string]any) error {
	msgDataJson, err := rdb().HGet(ctx, "chat_history_entries", CHEId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromJson[map[string]any](msgDataJson)

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "chat_history_entries", []string{CHEId, helpers.ToJson(msgData)}).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
