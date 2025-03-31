package realtimeService

import (
	"i9lyfe/src/appTypes"

	"github.com/zishang520/socket.io/v2/socket"
)

var sio *socket.Server

func SetIO(io *socket.Server) {
	sio = io
}

func HandleSocket(cliSocket *socket.Socket) {
	clientUsername := cliSocket.Data().(appTypes.ClientUser).Username

	cliSocket.Emit("greeting", "your're welcome, "+clientUsername)

	cliSocket.On("disconnect", func(args ...any) {
		cliSocket.Emit("greeting", "bye bye, "+clientUsername)
	})
}

func BroadcastNewPost(postId, ownerUsername string) {
	sio.FetchSockets()(func(rs []*socket.RemoteSocket, err error) {
		for _, cliSock := range rs {
			clientUsername := cliSock.Data().(appTypes.ClientUser).Username

			// get post
			postData := "working on it, " + clientUsername

			cliSock.Emit("new post", postData)
		}
	})
}
