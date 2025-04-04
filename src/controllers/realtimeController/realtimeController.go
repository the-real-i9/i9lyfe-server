package realtimeController

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/realtimeService"
	"log"

	"github.com/gofiber/contrib/websocket"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	realtimeService.AllClientSockets.Store(clientUser.Username, c)

	defer func() {
		realtimeService.AllClientSockets.Delete(clientUser.Username)
	}()

	go serverStream(ctx, clientUser.Username, c)

	var body struct {
		Event string `json:"event"`
		Data  any    `json:"data"`
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			return
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			return
		}
	}

})

func serverStream(ctx context.Context, clientUsername string, c *websocket.Conn) {

}
