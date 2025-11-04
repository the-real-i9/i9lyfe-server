package realtimeService

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"

	"github.com/gofiber/contrib/websocket"
)

func SendEventMsg(toUser string, msg appTypes.ServerEventMsg) {
	ctx := context.Background()

	// push to the queue for safety, in case the user
	// goes offline just before we send the message
	_, err := rdb.RPush(ctx, fmt.Sprintf("user_event_msgs_queue:%s", toUser), helpers.ToJson(msg)).Result()
	if err != nil {
		helpers.LogError(err)
		return
	}

	if userPipe, ok := AllClientSockets.Load(toUser); ok {
		pipe := userPipe.(*websocket.Conn)

		if err := pipe.WriteJSON(msg); err != nil {
			helpers.LogError(err)
			return
		}

		err := rdb.RPop(ctx, fmt.Sprintf("user_event_msgs_queue:%s", toUser)).Err()
		if err != nil {
			helpers.LogError(err)
		}
	}
}

func AddPipe(ctx context.Context, clientUsername string, pipe *websocket.Conn) {
	for {
		msgJson, err := rdb.LPop(ctx, fmt.Sprintf("user_event_msgs_queue:%s", clientUsername)).Result()
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				helpers.LogError(err)
				return
			}
			break
		}

		if err := pipe.WriteMessage(websocket.TextMessage, []byte(msgJson)); err != nil {
			helpers.LogError(err)
		}
	}

	AllClientSockets.Store(clientUsername, pipe)
}

func RemovePipe(clientUsername string) {
	AllClientSockets.Delete(clientUsername)
}
