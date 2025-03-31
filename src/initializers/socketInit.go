package initializers

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/services/realtimeService"
	"time"

	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

func InitSocket() (*socket.Server, *socket.ServerOptions) {
	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	// c.SetConnectionStateRecovery(&socket.ConnectionStateRecovery{})
	// c.SetAllowEIO3(true)
	c.SetPingInterval(300 * time.Millisecond)
	c.SetPingTimeout(200 * time.Millisecond)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(1000 * time.Millisecond)
	c.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	socketio := socket.NewServer(nil, nil)

	socketio.Use(authMiddlewares.UserAuthSocket)

	realtimeService.SetIO(socketio)

	socketio.Use(func(cliSocket *socket.Socket, next func(*socket.ExtendedError)) {
		// ...
		// realtime service
		realtimeService.HandleSocket(cliSocket)

		next(nil)
	})

	return socketio, c
}
