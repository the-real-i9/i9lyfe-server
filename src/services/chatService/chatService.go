package chatService

import (
	"context"
	"i9lyfe/src/appTypes/UITypes"
	chat "i9lyfe/src/models/chatModel"
)

func GetChats(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.ChatSnippet, error) {
	chats, err := chat.MyChats(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return chats, nil
}

// not implemented
func DeleteChat(ctx context.Context, clientUsername, partnerUsername string) (any, error) {
	done, err := chat.Delete(ctx, clientUsername, partnerUsername)
	if err != nil {
		return nil, err
	}

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	history, err := chat.History(ctx, clientUsername, partnerUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return history, nil
}
