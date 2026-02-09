package chatService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	chat "i9lyfe/src/models/chatModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetChats(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.ChatSnippet, error) {
	chats, err := chat.MyChats(ctx, clientUsername, limit, cursor)
	if err != nil {
		return nil, err
	}

	return chats, nil
}

// not implemented
func DeleteChat(ctx context.Context, clientUsername, partnerUsername string) (any, error) {
	done, err := chat.Delete(ctx, clientUsername, partnerUsername)
	if err != nil {
		return nil, err
	}

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	history, err := chat.History(ctx, clientUsername, partnerUsername, limit, cursor)
	if err != nil {
		return nil, err
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

	if newMessage.Id != "" {
		go func(msgData chat.NewMessageT) {
			uisender, err := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)
			if err != nil {
				return
			}

			uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)
			msgData.Sender = uisender

			cloudStorageService.MessageMediaCloudNameToUrl(msgData.Content)

			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: new message",
				Data:  msgData,
			})
		}(newMessage)

		go func(msgData chat.NewMessageT) {
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
	done, err := chat.AckMsgDelivered(ctx, clientUsername, partnerUsername, msgId, at)
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
	done, err := chat.AckMsgRead(ctx, clientUsername, partnerUsername, msgId, at)
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
	rxnToMessage, err := chat.ReactToMsg(ctx, clientUsername, partnerUsername, msgId, emoji, at)
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

		go func(rxnData chat.RxnToMessageT) {
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
	CHEId, err := chat.RemoveMsgReaction(ctx, clientUsername, partnerUsername, msgId)
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

	done, err := chat.DeleteMsg(ctx, clientUsername, partnerUsername, msgId, deleteFor, at)
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

/* ------------ */

type AuthMediaDataT struct {
	UploadUrl      string `json:"uploadUrl"`
	MediaCloudName string `json:"mediaCloudName"`
}

func AuthorizeUpload(ctx context.Context, msgType, mediaMIME string) (AuthMediaDataT, error) {
	var res AuthMediaDataT

	mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := cloudStorageService.GetUploadUrl(mediaCloudName, mediaMIME)
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

		mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s-%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString(), which[blurPlch0_actual1])

		url, err := cloudStorageService.GetUploadUrl(mediaCloudName, mime)
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
