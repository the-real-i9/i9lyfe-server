package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/chatService"
	"i9lyfe/src/services/chatService/chatMessageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/realtimeService"
	"i9lyfe/src/services/userService"
	"log"

	"github.com/gofiber/contrib/websocket"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	realtimeService.AllClientSockets.Store(clientUser.Username, c)

	go userService.GoOnline(context.Background(), clientUser.Username)

	eventStreamService.Subscribe(clientUser.Username, c)

	var w_err error

	for {
		var body rtActionBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
			continue
		}

		switch body.Action {
		case "start receiving post updates":
			realtimeService.PostUpdateSubscribers.Store(clientUser.Username, c)
		case "stop receiving post updates":
			realtimeService.PostUpdateSubscribers.Delete(clientUser.Username)
		case "start receiving comment updates":
			realtimeService.CommentUpdateSubscribers.Store(clientUser.Username, c)
		case "stop receiving comment updates":
			realtimeService.CommentUpdateSubscribers.Delete(clientUser.Username)
		case "chat: send message":
			var data sendChatMsgAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.SendMessage(ctx, clientUser.Username, data.PartnerUsername, data.ReplyTargetMsgId, data.IsReply, data.Msg, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: get history":
			var data getChatHistoryAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatService.GetChatHistory(ctx, clientUser.Username, data.PartnerUsername, data.Offset)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: ack message delivered":
			var data ackChatMsgDeliveredAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.AckMsgDelivered(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: ack message read":
			var data ackChatMsgReadAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.AckMsgRead(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: react to message":
			var data reactToChatMsgAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.ReactToMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.Reaction, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: remove reaction to message":
			var data removeReactionToChatMsgAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.RemoveReactionToMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))
		case "chat: delete message":
			var data deleteChatMsgAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			res, err := chatMessageService.DeleteMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.DeleteFor)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Action))

		default:
			w_err = c.WriteJSON(helpers.WSErrReply(fmt.Errorf("invalid event: %s", body.Action), body.Action))
			continue
		}
	}

	realtimeService.AllClientSockets.Delete(clientUser.Username)

	go userService.GoOffline(context.Background(), clientUser.Username)

	eventStreamService.Unsubscribe(clientUser.Username)
})
