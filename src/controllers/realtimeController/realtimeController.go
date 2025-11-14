package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/chatService"
	"i9lyfe/src/services/chatService/chatMessageService"
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
			var data subToUserPresenceAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			for _, tu := range data.Usernames {
				ctx, cancel := context.WithCancel(ctx)

				realtimeService.SubscribeToUserPresence(ctx, clientUser.Username, tu, cancel)

				cancelUserPresenceSub[tu] = cancel
			}
		case "unsubscribe from user presence change":
			var data unsubFromUserPresenceAcd

			helpers.ToStruct(body.Data, &data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Action))
				continue
			}

			for _, tu := range data.Usernames {
				if cancel, ok := cancelUserPresenceSub[tu]; ok {
					cancel()
				}

				delete(cancelUserPresenceSub, tu)
			}
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

			if data.Limit == 0 {
				data.Limit = 50
			}

			res, err := chatService.GetChatHistory(ctx, clientUser.Username, data.PartnerUsername, data.Limit, data.Cursor)
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

	go userService.GoOffline(context.Background(), clientUser.Username)

	realtimeService.RemovePipe(clientUser.Username)
})
