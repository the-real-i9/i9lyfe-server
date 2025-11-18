package chatMessageService

import (
	"context"
	"encoding/json"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
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

	if newMessage.Id != "" {
		go func(msgData chatMessage.NewMessageT) {
			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: new message",
				Data:  msgData,
			})
		}(newMessage)

		go func(msgData chatMessage.NewMessageT) {
			msgData.Sender = clientUsername

			eventStreamService.QueueNewMessageEvent(eventTypes.NewMessageEvent{
				FirstFromUser: msgData.FirstFromUser,
				FirstToUser:   msgData.FirstToUser,
				FromUser:      clientUsername,
				ToUser:        partnerUsername,
				CHEId:         msgData.Id,
				MsgData:       helpers.ToJson(msgData),
			})
		}(newMessage)
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
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    msgId,
			Ack:      "delivered",
			At:       at,
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
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    msgId,
			Ack:      "read",
			At:       at,
		})
	}

	return done, nil
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (any, error) {
	rxnToMessage, err := chatMessage.ReactTo(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := rxnToMessage.CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction",
			Data: map[string]any{
				"partner_username": clientUsername,
				"to_msg_id":        msgId,
				"reaction": UITypes.MsgReaction{
					Emoji:   emoji,
					Reactor: rxnToMessage.Reactor,
				},
			},
		})

		go func(rxnData chatMessage.RxnToMessageT) {
			rxnData.Reactor = clientUsername

			// push CHE id to each user's chat history
			// store Rxn data to ToMsg reactions
			eventStreamService.QueueNewMsgReactionEvent(eventTypes.NewMsgReactionEvent{
				FromUser: clientUsername,
				ToUser:   partnerUsername,
				CHEId:    rxnData.CHEId,
				RxnData:  helpers.ToJson(rxnData),
				ToMsgId:  rxnData.ToMsgId,
				Emoji:    rxnData.Emoji,
			})
		}(rxnToMessage)
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
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			ToMsgId:  msgId,
			CHEId:    CHEId,
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
	}

	if done {
		go eventStreamService.QueueMsgDeletionEvent(eventTypes.MsgDeletionEvent{
			CHEId: msgId,
			For:   deleteFor,
		})

	}

	return done, nil
}
