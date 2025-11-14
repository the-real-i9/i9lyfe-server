package chatService

import (
	"context"
	chat "i9lyfe/src/models/chatModel"
)

func GetChats(ctx context.Context, clientUsername string, limit int, cursor float64) (any, error) {
	chats, err := chat.MyChats(ctx, clientUsername, limit, cursor)
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

	return true, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) (any, error) {
	history, err := chat.History(ctx, clientUsername, partnerUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return history, nil
}
