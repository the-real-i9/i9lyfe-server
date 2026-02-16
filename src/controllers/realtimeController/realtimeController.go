package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/controllers/chatControllers"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/realtimeService"
	"i9lyfe/src/services/userService"
	"log"

	"github.com/gofiber/contrib/websocket"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	go userService.GoOnline(context.Background(), clientUser.Username)

	realtimeService.AddPipe(ctx, clientUser.Username, c)

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

		var (
			lcmCtx       context.Context
			cancelLcmSub context.CancelFunc
		)

		cancelUserPresenceSub := make(map[string]context.CancelFunc)

		switch body.Action {
		case "subscribe to live content metrics":
			lcmCtx, cancelLcmSub = context.WithCancel(ctx)

			realtimeService.SubscribeToLiveContentMetrics(lcmCtx, clientUser.Username, cancelLcmSub)
		case "unsubscribe from live content metrics":
			cancelLcmSub()
		case "subscribe to user presence change":

			data := helpers.FromBtMsgPack[subToUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				ctx, cancel := context.WithCancel(ctx)

				realtimeService.SubscribeToUserPresence(ctx, clientUser.Username, tu, cancel)

				cancelUserPresenceSub[tu] = cancel
			}
		case "unsubscribe from user presence change":

			data := helpers.FromBtMsgPack[unsubFromUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				if cancel, ok := cancelUserPresenceSub[tu]; ok {
					cancel()
				}

				delete(cancelUserPresenceSub, tu)
			}
		case "chat: send message":
			res, err := chatControllers.SendMessage(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))

		case "chat: ack message delivered":
			res, err := chatControllers.AckMsgDelivered(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(res, body.Action)))
		case "chat: ack message read":
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
			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(fmt.Errorf("invalid event: %s", body.Action), body.Action)))
			continue
		}
	}

	go userService.GoOffline(context.Background(), clientUser.Username)

	realtimeService.RemovePipe(clientUser.Username)
})
