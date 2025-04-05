package realtimeController

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/messageBrokerService"
	"i9lyfe/src/services/realtimeService"
	"i9lyfe/src/services/userService"
	"io"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/segmentio/kafka-go"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	realtimeService.AllClientSockets.Store(clientUser.Username, c)

	go userService.GoOnline(ctx, clientUser.Username)

	r := messageBrokerService.ConsumeTopic(fmt.Sprintf("user-%s-alerts", clientUser.Username))

	go serverStream(c, r)

	var w_err error

	for {
		var body clientMessageBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.WSErrReply(val_err, body.Event))
			continue
		}

		switch body.Event {
		case "start receiving post updates":
			realtimeService.PostUpdateSubscribers.Store(clientUser.Username, c)
		case "stop receiving post updates":
			realtimeService.PostUpdateSubscribers.Delete(clientUser.Username)
		case "send message":
		case "get chat history":
		case "ack message delivered":
		case "ack message read":
		case "react to message":
		case "remove reaction to message":
		case "delete message":
		}
	}

	go func(r *kafka.Reader, clientUsername string) {
		realtimeService.AllClientSockets.Delete(clientUsername)

		userService.GoOffline(context.Background(), clientUsername)

		if err := r.Close(); err != nil {
			log.Println("failed to close reader:", err)
		}
	}(r, clientUser.Username)
})

func serverStream(c *websocket.Conn, r *kafka.Reader) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("realtimeController.go: serverStream: r.ReadMessage:", err)
			}
			break
		}

		var msg any
		json.Unmarshal(m.Value, &msg)

		c.WriteJSON(msg)
	}
}
