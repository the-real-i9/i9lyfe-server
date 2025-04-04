package chatService

import (
	"context"
	"i9lyfe/src/appGlobals"
	chat "i9lyfe/src/models/chatModel"
	"time"
)

func GetChats(ctx context.Context, clientUsername string, limit int, offset int64) ([]any, error) {
	chats, err := chat.All(ctx, clientUsername, limit, time.UnixMilli(offset).UTC())
	if err != nil {
		return nil, err
	}

	return chats, nil
}

func DeleteChat(ctx context.Context, clientUsername, partnerUsername string) (any, error) {
	err := chat.Delete(ctx, clientUsername, partnerUsername)
	if err != nil {
		return nil, err
	}

	return appGlobals.OprSucc, nil
}
