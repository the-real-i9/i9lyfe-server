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
func DeleteChat(ctx context.Context, clientUsername, partnerUsername string) (bool, error) {
	done, err := chat.Delete(ctx, clientUsername, partnerUsername)
	if err != nil {
		return false, err
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

	if newMessage.Id == "" {
		return nil, nil
	}

	now := float64(time.Now().UnixMicro())

	go func(msg chat.NewMessageT) {
		uisender, _ := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)

		uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)

		cloudStorageService.MessageMediaCloudNameToUrl(msg.Content)

		msgUI := UITypes.ChatHistoryEntry{CHEType: msg.CHEType, Id: msg.Id, Content: msg.Content, DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt, Sender: uisender, ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: msg.Snum + now}

		if newMessage.FirstToUser {
			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "new chat",
				Data: map[string]any{
					"chat":    UITypes.ChatSnippet{PartnerUser: uisender, UnreadMC: 1, Cursor: msg.Snum + now},
					"history": []UITypes.ChatHistoryEntry{msgUI},
				},
			})

			return
		}

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: new message",
			Data:  msgUI,
		})
	}(newMessage)

	go eventStreamService.QueueNewMessageEvent(eventTypes.NewMessageEvent{
		FirstFromUser: newMessage.FirstFromUser,
		FirstToUser:   newMessage.FirstToUser,
		FromUser:      clientUsername,
		ToUser:        partnerUsername,
		CHEId:         newMessage.Id,
		MsgData:       helpers.ToJson(newMessage),
		Score:         newMessage.Snum + now,
	})

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Snum + now}, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername string, msgIdList []string, at int64) (bool, error) {
	done, err := chat.AckMsgDelivered(ctx, clientUsername, partnerUsername, msgIdList, at)
	if err != nil {
		return false, err
	}

	if done {
		now := float64(time.Now().UnixMicro())

		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: messages delivered",
			Data: map[string]any{
				"chat_partner":    clientUsername,
				"msg_id_list":     msgIdList,
				"delivered_at":    at,
				"new_chat_cursor": now,
			},
		})

		go eventStreamService.QueueMsgsAckEvent(eventTypes.MsgsAckEvent{
			FromUser:  clientUsername,
			ToUser:    partnerUsername,
			CHEIdList: msgIdList,
			Ack:       "delivered",
			At:        at,
			Score:     now,
		})
	}

	return done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername string, msgIdList []string, at int64) (bool, error) {
	done, err := chat.AckMsgRead(ctx, clientUsername, partnerUsername, msgIdList, at)
	if err != nil {
		return false, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: messages read",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_id_list":  msgIdList,
				"read_at":      at,
			},
		})

		go eventStreamService.QueueMsgsAckEvent(eventTypes.MsgsAckEvent{
			FromUser:  clientUsername,
			ToUser:    partnerUsername,
			CHEIdList: msgIdList,
			Ack:       "read",
			At:        at,
		})
	}

	return done, nil
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (map[string]any, error) {
	rxnToMessage, err := chat.ReactToMsg(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	if rxnToMessage.CHEId == "" {
		return nil, nil
	}

	now := float64(time.Now().UnixMicro())

	go func(rxnData chat.RxnToMessageT) {
		uireactor, _ := cache.GetUser[UITypes.MsgReactor](context.Background(), clientUsername)

		uireactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uireactor.ProfilePicUrl)

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"che":          UITypes.ChatHistoryEntry{CHEType: rxnData.CHEType, Reactor: clientUsername, Emoji: rxnData.Emoji, Cursor: rxnData.Snum + now},
				"msg_reaction": map[string]any{
					"msg_id": msgId,
					"reaction": UITypes.MsgReaction{
						Emoji:   emoji,
						Reactor: uireactor,
					},
				},
			},
		})
	}(rxnToMessage)

	go func(rxnData chat.RxnToMessageT) {
		eventStreamService.QueueNewMsgReactionEvent(eventTypes.NewMsgReactionEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    rxnData.CHEId,
			RxnData:  helpers.ToJson(rxnData),
			ToMsgId:  rxnData.ToMsgId,
			Emoji:    rxnData.Emoji,
			Score:    rxnData.Snum + now,
		})
	}(rxnToMessage)

	return map[string]any{"che_cursor": rxnToMessage.Snum + now}, nil
}

func RemoveReactionToMsg(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	CHEId, err := chat.RemoveMsgReaction(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return false, err
	}

	done := CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "chat: message reaction removed",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_id":       msgId,
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

func DeleteMsg(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (bool, error) {
	at := time.Now().UnixMilli()

	done, err := chat.DeleteMsg(ctx, clientUsername, partnerUsername, msgId, deleteFor, at)
	if err != nil {
		return false, err
	}

	if done {
		if deleteFor == "everyone" {
			go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "chat: message deleted",
				Data: map[string]any{
					"chat_partner": clientUsername,
					"msg_id":       msgId,
				},
			})
		}

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
