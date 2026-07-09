package chatService

import (
	"context"
	"fmt"

	chat "i9lyfe/src/domain/chat/chatDBM"
	"i9lyfe/src/services/mediaStorageService"
	"i9lyfe/src/services/sseService"
	"i9lyfe/src/types/UITypes"
	"i9lyfe/src/types/appTypes"

	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

func GetChats(ctx context.Context, clientUsername string, limit int64, cursor int64) ([]*UITypes.ChatSnippet, error) {
	chats, err := chat.MyChats(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, c := range chats {
		c.PartnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(c.PartnerUser["profile_pic_url"].(string))
	}

	return chats, nil
}

func DeleteChat(ctx context.Context, clientUsername, partnerUsername string) (bool, error) {
	done, err := chat.Delete(ctx, clientUsername, partnerUsername)
	if err != nil {
		return false, err
	}

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int64, cursor float64) ([]*UITypes.ChatHistoryEntry, error) {
	history, err := chat.History(ctx, clientUsername, partnerUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, h := range history {
		if h.CHEType == "message" {
			h.Content = mediaStorageService.MessageMediaCloudNameToUrl(h.Content)
		}
	}

	return history, nil
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, replyTargetMsgId string, isReply bool, msgContentJson string, at int64) (map[string]any, error) {
	var (
		newMessage chat.NewMessageT
		err        error
	)

	if !isReply {
		newMessage, err = chat.SendMessage(ctx, clientUsername, partnerUsername, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = chat.ReplyMessage(ctx, clientUsername, partnerUsername, replyTargetMsgId, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	}

	if newMessage.Id == "" {
		return nil, nil
	}

	go func(msg chat.NewMessageT) {
		msg.Sender["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(msg.Sender["profile_pic_url"].(string))

		UImsg := UITypes.ChatHistoryEntry{
			CHEType: msg.CHEType, Id: msg.Id,
			Content:        mediaStorageService.MessageMediaCloudNameToUrl(msg.Content),
			DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt, Sender: msg.Sender,
			ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: msg.Cursor,
		}

		sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: new che: message",
			Data:  UImsg,
		})
	}(newMessage)

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Cursor}, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername string, msgIds []string, at int64) (bool, error) {
	done, err := chat.AckMsgDelivered(ctx, clientUsername, partnerUsername, msgIds, at)
	if err != nil {
		return false, err
	}

	if done {
		go sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: messages delivered",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_ids":      msgIds,
				"delivered_at": at,
			},
		})
	}

	return done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername string, msgIds []string, at int64) (bool, error) {
	done, err := chat.AckMsgRead(ctx, clientUsername, partnerUsername, msgIds, at)
	if err != nil {
		return false, err
	}

	if done {
		go sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: messages read",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_ids":      msgIds,
				"read_at":      at,
			},
		})
	}

	return done, nil
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (bool, error) {
	rxnToMessage, err := chat.ReactToMsg(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return false, err
	}

	done := rxnToMessage.CHEId != ""

	if done {
		go func(rxnToMessage chat.RxnToMessageT, clientUsername, partnerUsername string) {
			rxnToMessage.Reactor["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(rxnToMessage.Reactor["profile_pic_url"].(string))

			sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: new che: reaction",
				Data:  UITypes.ChatHistoryEntry{Id: rxnToMessage.CHEId, CHEType: rxnToMessage.CHEType, Reactor: rxnToMessage.Reactor, Emoji: rxnToMessage.Emoji, ToMsg: rxnToMessage.ToMsg, Cursor: rxnToMessage.Cursor},
			})
		}(rxnToMessage, clientUsername, partnerUsername)
	}

	return done, nil
}

func RemoveReactionToMsg(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	CHEId, err := chat.RemoveMsgReaction(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return false, err
	}

	done := CHEId != ""

	if done {
		go sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction removed",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"che_id":       CHEId,
				"msg_id":       msgId,
			},
		})
	}

	return done, nil
}

func DeleteMsg(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (bool, error) {
	at := time.Now().UnixMilli()

	done, err := chat.DeleteMsg(ctx, clientUsername, partnerUsername, msgId, deleteFor, at)
	if err != nil {
		return false, err
	}

	if done {
		if deleteFor == "everyone" {
			go sseService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: message deleted",
				Data: map[string]any{
					"chat_partner": clientUsername,
					"msg_id":       msgId,
				},
			})
		}
	}

	return done, nil
}

/* ------------ */

type AuthMediaDataT struct {
	UploadUrl      string `msgpack:"uploadUrl"`
	MediaCloudName string `msgpack:"mediaCloudName"`
}

func AuthorizeUpload(ctx context.Context, msgType, mediaMIME string) (AuthMediaDataT, error) {
	var res AuthMediaDataT

	mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s", msgType, time.Now().Year(), time.Now().Month(), utils.UUIDv4())

	url, err := mediaStorageService.GetUploadUrl(mediaCloudName, mediaMIME)
	if err != nil {
		return AuthMediaDataT{}, fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.MediaCloudName = mediaCloudName

	return res, nil
}

func AuthorizeVisualUpload(ctx context.Context, msgType string, mediaMIME [2]string) (AuthMediaDataT, error) {
	var res AuthMediaDataT

	for blurPlch0_actual1, mime := range mediaMIME {

		which := [2]string{"blur_placeholder", "actual"}

		mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s-%s", msgType, time.Now().Year(), time.Now().Month(), utils.UUIDv4(), which[blurPlch0_actual1])

		url, err := mediaStorageService.GetUploadUrl(mediaCloudName, mime)
		if err != nil {
			return AuthMediaDataT{}, fiber.ErrInternalServerError
		}

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += "blur_placeholder:"
			res.MediaCloudName += "blur_placeholder:"
		} else {
			res.UploadUrl += "actual:"
			res.MediaCloudName += "actual:"
		}

		res.UploadUrl += url
		res.MediaCloudName += mediaCloudName

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += " "
			res.MediaCloudName += " "
		}
	}

	return res, nil
}
