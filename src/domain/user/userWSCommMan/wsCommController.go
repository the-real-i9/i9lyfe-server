package userWSCommMan

import (
	"context"
	"fmt"

	"i9lyfe/src/domain/chat/chatControllers"
	"i9lyfe/src/domain/user/userService"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/pubsubService"
	"i9lyfe/src/services/sseService"
	"i9lyfe/src/types/appTypes"

	"log"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	go userService.GoOnline(context.Background(), clientUser.Username)

	sseService.AddUserSocket(ctx, clientUser.Username, c)

	var w_err error

	for {

		if w_err != nil {
			log.Println(w_err)
			break
		}

		MSG_TYPE, msgPackBt, r_err := c.ReadMessage()
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if MSG_TYPE != websocket.BinaryMessage {
			w_err = c.WriteMessage(websocket.BinaryMessage, helpers.ToBtMsgPack(fmt.Errorf("unexpected message type")))
			continue
		}

		body := helpers.FromBtMsgPack[rtActionBody](msgPackBt)

		if err := body.Validate(); err != nil {
			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
			continue
		}

		var unsubLcm func()

		unsubUserPresenceChange := make(map[string]func())

		switch body.Action {
		case "subscribe to live content metrics":

			unsubLcm = pubsubService.SubscribeToLiveContentMetrics(ctx, clientUser.Username)

		case "unsubscribe from live content metrics":
			unsubLcm()
		case "subscribe to user presence change":

			data := helpers.FromBtMsgPack[subToUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				unsubUserPresenceChange[tu] = pubsubService.SubscribeToUserPresence(ctx, clientUser.Username, tu)
			}
		case "unsubscribe from user presence change":

			data := helpers.FromBtMsgPack[unsubFromUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				if unsubUpc, ok := unsubUserPresenceChange[tu]; ok {
					unsubUpc()
				}

				delete(unsubUserPresenceChange, tu)
			}
		case "chat: send message":
			res, err := chatControllers.SendMessage(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))

		case "chat: ack messages delivered":
			res, err := chatControllers.AckMsgDelivered(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: ack messages read":
			res, err := chatControllers.AckMsgRead(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: get history":
			res, err := chatControllers.GetChatHistory(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: react to message":
			res, err := chatControllers.ReactToMsg(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: remove reaction to message":
			res, err := chatControllers.RemoveReactionToMsg(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: delete message":
			res, err := chatControllers.DeleteMsg(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))

		default:
			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(fiber.NewErrorf(fiber.StatusBadRequest, "invalid event: %s", body.Action), body.Action)))
			continue
		}
	}

	go userService.GoOffline(context.Background(), clientUser.Username)

	sseService.RemoveUserSocket(clientUser.Username)
})
