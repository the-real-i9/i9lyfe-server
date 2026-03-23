package sseService

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"sync"

	"github.com/gofiber/contrib/v3/websocket"
)

var AllClientSockets = &sync.Map{}

func SendEventMsg(toUser string, msg appTypes.ServerEventMsg) {
	if userSock, ok := AllClientSockets.Load(toUser); ok {
		sock := userSock.(*websocket.Conn)

		if err := sock.WriteMessage(websocket.BinaryMessage, helpers.ToBtMsgPack(msg)); err != nil {
			helpers.LogError(err)
		}

		return
	}
}

func GetUserSocket(username string) (any, bool) {
	return AllClientSockets.Load(username)
}

func AddUserSocket(ctx context.Context, clientUsername string, sock *websocket.Conn) {
	AllClientSockets.Store(clientUsername, sock)
}

func RemoveUserSocket(clientUsername string) {
	AllClientSockets.Delete(clientUsername)
}
