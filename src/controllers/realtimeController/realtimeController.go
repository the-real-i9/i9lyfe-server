package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/chatMessageTypes"
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
		case "start receiving comment updates":
			realtimeService.CommentUpdateSubscribers.Store(clientUser.Username, c)
		case "stop receiving comment updates":
			realtimeService.CommentUpdateSubscribers.Delete(clientUser.Username)
		case "chat: send message: text":
			var msg chatMessageTypes.Text

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendTextMessage(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: send message: voice":
			var msg chatMessageTypes.Voice

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendVoiceMessage(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: send message: photo":
			var msg chatMessageTypes.Photo

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendPhoto(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: send message: video":
			var msg chatMessageTypes.Video

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendVideo(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: send message: audio":
			var msg chatMessageTypes.Audio

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendAudio(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: send message: file":
			var msg chatMessageTypes.File

			helpers.ToStruct(body.Data, &msg)

			if err := msg.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.SendFile(ctx, clientUser.Username, msg)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: get history":
			var data getChatHistoryEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatService.GetChatHistory(ctx, clientUser.Username, data.PartnerUsername, data.Offset)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: ack message delivered":
			var data ackChatMsgDeliveredEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.AckMsgDelivered(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: ack message read":
			var data ackChatMsgReadEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.AckMsgRead(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: react to message":
			var data reactToChatMsgEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.ReactToMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.Reaction, data.At)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: remove reaction to message":
			var data removeReactionToChatMsgEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.RemoveReactionToMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))
		case "chat: delete message":
			var data deleteChatMsgEvd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			res, err := chatMessageService.DeleteMsg(ctx, clientUser.Username, data.PartnerUsername, data.MsgId, data.DeleteFor)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			w_err = c.WriteJSON(helpers.WSReply(res, body.Event))

		default:
			w_err = c.WriteJSON(helpers.WSErrReply(fmt.Errorf("invalid event: %s", body.Event), body.Event))
			continue
		}
	}

	realtimeService.AllClientSockets.Delete(clientUser.Username)

	go userService.GoOffline(context.Background(), clientUser.Username)

	eventStreamService.Unsubscribe(clientUser.Username)
})
