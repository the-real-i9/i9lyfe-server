package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/realtimeService"
	"sync"

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
		Event string         `json:"event"`
		Data  map[string]any `json:"data"`
	}

	var w_err error

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			fmt.Println(r_err)
			return
		}
	}

})

func serverStream(ctx context.Context, clientUsername string, c *websocket.Conn) {

}
