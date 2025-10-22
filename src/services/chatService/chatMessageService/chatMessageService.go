package chatMessageService

import (
	"context"
	"encoding/json"
	"i9lyfe/src/appTypes"
	chatMessage "i9lyfe/src/models/chatModel/chatMessageModel"
	"i9lyfe/src/services/eventStreamService"
	"log"
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
		log.Println("chatMessageService.go: SendMessage: json.Marshal:", err)
		return nil, fiber.ErrInternalServerError
	}

	var newMessage chatMessage.NewMessage

	if !isReply {
		newMessage, err = chatMessage.Send(ctx, clientUsername, partnerUsername, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = chatMessage.Reply(ctx, clientUsername, partnerUsername, replyTargetMsgId, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	}

	if newMessage.PartnerData != nil {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  newMessage.PartnerData,
		})
	}

	return newMessage.ClientData, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, at int64) (any, error) {
	done, err := chatMessage.AckDelivered(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message delivered",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"delivered_at":     at,
			},
		})
	}

	return done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername, msgId string, at int64) (any, error) {
	done, err := chatMessage.AckRead(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message read",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"read_at":          at,
			},
		})
	}

	return done, nil
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, reaction string, at int64) (any, error) {
	rxnToMessage, err := chatMessage.ReactTo(ctx, clientUsername, partnerUsername, msgId, reaction, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if rxnToMessage.PartnerData != nil {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message reaction",
			Data:  rxnToMessage.PartnerData,
		})
	}

	return rxnToMessage.ClientData, nil
}

func RemoveReactionToMsg(ctx context.Context, clientUsername, partnerUsername, msgId string) (any, error) {
	done, err := chatMessage.RemoveReaction(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message reaction removed",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}

func DeleteMsg(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (any, error) {
	done, err := chatMessage.Delete(ctx, clientUsername, partnerUsername, msgId, deleteFor)
	if err != nil {
		return nil, err
	}

	if done && deleteFor == "everyone" {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message deleted",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}
