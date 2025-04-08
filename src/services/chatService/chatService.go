package chatService

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	chat "i9lyfe/src/models/chatModel"
)

func GetChats(ctx context.Context, clientUsername string, limit int, offset int64) ([]any, error) {
	chats, err := chat.All(ctx, clientUsername, limit, helpers.OffsetTime(offset))
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

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, offset int64) ([]any, error) {
	history, err := chat.History(ctx, clientUsername, partnerUsername, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return history, nil
}
