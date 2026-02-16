package cache

import (
	"context"
	"i9lyfe/src/helpers"
	"maps"
)

func UpdateUser(ctx context.Context, username string, updateKVMap map[string]any) error {
	userDataMsgPack, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	userData := helpers.FromMsgPack[map[string]any](userDataMsgPack)

	maps.Copy(userData, updateKVMap)

	err = rdb().HSet(ctx, "users", username, helpers.ToMsgPack(userData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateMessageDelivery(ctx context.Context, msgId string, updateKVMap map[string]any) error {
	msgDataMsgPack, err := rdb().HGet(ctx, "chat_history_entries", msgId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromMsgPack[map[string]any](msgDataMsgPack)

	// if a client skips the "delivered" ack, and acks "read"
	// it means the message is delivered and read at the same time
	if updateKVMap["read_at"] != nil && msgData["delivered_at"] == nil {
		msgData["delivered_at"] = updateKVMap["read_at"]
	}

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "chat_history_entries", msgId, helpers.ToMsgPack(msgData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
