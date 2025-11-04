package chatMessageService

import (
	"context"
	"encoding/json"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	chatMessage "i9lyfe/src/models/chatModel/chatMessageModel"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendMessage(ctx context.Context, clientUsername, partnerUsername, replyTargetMsgId string, isReply bool, msgContent *appTypes.MsgContent, at int64) (map[string]any, error) {

	err := uploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson, err := json.Marshal(*msgContent)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	var newMessage chatMessage.NewMessageT

	if !isReply {
		newMessage, err = chatMessage.Send(ctx, clientUsername, partnerUsername, string(msgContentJson), at)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = chatMessage.Reply(ctx, clientUsername, partnerUsername, replyTargetMsgId, string(msgContentJson), at)
		if err != nil {
			return nil, err
		}
	}

	msgDataMap := helpers.StructToMap(newMessage)

	if newMessage.Id != "" {
		go func(msgData map[string]any) {
			msgData["is_own"] = false

			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: new message",
				Data:  msgData,
			})
		}(msgDataMap)

		go func(msgData map[string]any) {
			delete(msgData, "sender")
			msgData["sender_username"] = clientUsername

			// store CHE data to cache, direct
			// push CHE id to each user's chat history
			eventStreamService.QueueNewMessageEvent(eventTypes.NewMessageEvent{
				FromUser: clientUsername,
				ToUser:   partnerUsername,
				CHEId:    newMessage.Id,
				MsgData:  msgData,
			})
		}(msgDataMap)
	}

	return map[string]any{"new_msg_id": newMessage.Id}, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, at int64) (any, error) {
	done, err := chatMessage.AckDelivered(ctx, clientUsername, partnerUsername, msgId, at)
	if err != nil {
		return nil, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message delivered",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"delivered_at":     at,
			},
		})

		go eventStreamService.QueueMsgAckEvent(eventTypes.MsgAckEvent{
			CHEId: msgId,
			Ack:   "delivered",
			At:    at,
		})
	}

	return done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername, msgId string, at int64) (any, error) {
	done, err := chatMessage.AckRead(ctx, clientUsername, partnerUsername, msgId, at)
	if err != nil {
		return nil, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message read",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"read_at":          at,
			},
		})

		go eventStreamService.QueueMsgAckEvent(eventTypes.MsgAckEvent{
			CHEId: msgId,
			Ack:   "read",
			At:    at,
		})
	}

	return done, nil
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (any, error) {
	rxnToMessage, err := chatMessage.ReactTo(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := rxnToMessage.Id != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction",
			Data: map[string]any{
				"partner_username": clientUsername,
				"reactor":          rxnToMessage.Reactor,
				"to_msg_id":        msgId,
				"emoji":            emoji,
				"at":               at,
			},
		})

		go func(rxnData map[string]any) {
			delete(rxnData, "reactor")
			rxnData["reactor_username"] = clientUsername

			// store CHE data to cache, direct
			// push CHE id to each user's chat history
			eventStreamService.QueueNewMsgReactionEvent(eventTypes.NewMsgReactionEvent{
				FromUser: clientUsername,
				ToUser:   partnerUsername,
				CHEId:    rxnToMessage.Id,
				RxnData:  rxnData,
			})
		}(helpers.StructToMap(rxnToMessage))
	}

	return done, nil
}

func RemoveReactionToMsg(ctx context.Context, clientUsername, partnerUsername, msgId string) (any, error) {
	CHEId, err := chatMessage.RemoveReaction(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return nil, err
	}

	done := CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction removed",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})

		go eventStreamService.QueueMsgReactionRemovedEvent(eventTypes.MsgReactionRemovedEvent{
			CHEId: CHEId,
		})
	}

	return done, nil
}

func DeleteMsg(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (any, error) {
	at := time.Now().UnixMilli()

	done, err := chatMessage.Delete(ctx, clientUsername, partnerUsername, msgId, deleteFor, at)
	if err != nil {
		return nil, err
	}

	if done && deleteFor == "everyone" {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message deleted",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})

		go eventStreamService.QueueMsgDeletionEvent(eventTypes.MsgDeletionEvent{
			CHEId: msgId,
			For:   deleteFor,
		})
	}

	return done, nil
}
