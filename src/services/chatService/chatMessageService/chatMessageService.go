package chatMessageService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/chatMessageTypes"
	"i9lyfe/src/helpers"
	chatMessage "i9lyfe/src/models/chatModel/chatMessageModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"log"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func SendTextMessage(ctx context.Context, clientUsername string, msg chatMessageTypes.Text) (any, error) {
	msgType := "text"

	msgProps, err := json.Marshal(msg.Props)
	if err != nil {
		log.Println("chatService.go: SendTextMessage:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
}

func SendVoiceMessage(ctx context.Context, clientUsername string, msg chatMessageTypes.Voice) (any, error) {
	msgType := "voice"

	voiceData := msg.Props.Data

	mediaMIME := mimetype.Detect(voiceData)
	mediaType := mediaMIME.String()
	mediaExt := mediaMIME.Extension()

	if !strings.HasPrefix((mediaType), "audio") {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid media type for voice message. expected audio/*, but found "+mediaType)
	}

	// upload the voice data in exchange for a url
	voiceUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("message_medias/user-%s-%s", clientUsername, msg.ToUser), voiceData, mediaExt)
	if err != nil {
		return nil, err
	}

	var msgPropsMap map[string]any

	helpers.StructToMap(msg.Props, &msgPropsMap)

	msgPropsMap["data_url"] = voiceUrl

	delete(msgPropsMap, "data")

	// marshal message props to string ([]byte) for db storage
	msgProps, err := json.Marshal(msgPropsMap)
	if err != nil {
		log.Println("chatService.go: SendVoiceMessage:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
}

func SendPhoto(ctx context.Context, clientUsername string, msg chatMessageTypes.Photo) (any, error) {
	msgType := "photo"

	photoData := msg.Props.Data

	mediaMIME := mimetype.Detect(photoData)
	mediaType := mediaMIME.String()
	mediaExt := mediaMIME.Extension()

	if !strings.HasPrefix((mediaType), "image") {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid media type for photo message. expected image/*, but found "+mediaType)
	}

	// upload photo data in exchange for a url
	photoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("message_medias/user-%s-%s", clientUsername, msg.ToUser), photoData, mediaExt)
	if err != nil {
		return nil, err
	}

	var msgPropsMap map[string]any

	helpers.StructToMap(msg.Props, &msgPropsMap)

	msgPropsMap["data_url"] = photoUrl

	delete(msgPropsMap, "data")

	// marshal message props to string ([]byte) for db storage
	msgProps, err := json.Marshal(msgPropsMap)
	if err != nil {
		log.Println("chatService.go: SendPhoto:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
}

func SendVideo(ctx context.Context, clientUsername string, msg chatMessageTypes.Video) (any, error) {
	msgType := "video"

	videoData := msg.Props.Data

	mediaMIME := mimetype.Detect(videoData)
	mediaType := mediaMIME.String()
	mediaExt := mediaMIME.Extension()

	if !strings.HasPrefix((mediaType), "video") {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid media type for video message. expected video/*, but found "+mediaType)
	}

	// upload the video data in exchange for a url
	videoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("message_medias/user-%s-%s", clientUsername, msg.ToUser), videoData, mediaExt)
	if err != nil {
		return nil, err
	}

	var msgPropsMap map[string]any

	helpers.StructToMap(msg.Props, &msgPropsMap)

	msgPropsMap["data_url"] = videoUrl

	delete(msgPropsMap, "data")

	// marshal message props to string ([]byte) for db storage
	msgProps, err := json.Marshal(msgPropsMap)
	if err != nil {
		log.Println("chatService.go: SendVideo:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
}

func SendAudio(ctx context.Context, clientUsername string, msg chatMessageTypes.Audio) (any, error) {
	msgType := "audio"

	audioData := msg.Props.Data

	mediaMIME := mimetype.Detect(audioData)
	mediaType := mediaMIME.String()
	mediaExt := mediaMIME.Extension()

	if !strings.HasPrefix((mediaType), "audio") {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid media type for audio message. expected audio/*, but found "+mediaType)
	}

	// upload the audio data in exchange for a url
	audioUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("message_medias/user-%s-%s", clientUsername, msg.ToUser), audioData, mediaExt)
	if err != nil {
		return nil, err
	}

	var msgPropsMap map[string]any

	helpers.StructToMap(msg.Props, &msgPropsMap)

	msgPropsMap["data_url"] = audioUrl

	delete(msgPropsMap, "data")

	// marshal message props to string ([]byte) for db storage
	msgProps, err := json.Marshal(msgPropsMap)
	if err != nil {
		log.Println("chatService.go: SendAudio:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
}

func SendFile(ctx context.Context, clientUsername string, msg chatMessageTypes.File) (any, error) {
	msgType := "file"

	fileData := msg.Props.Data

	mediaMIME := mimetype.Detect(fileData)
	mediaExt := mediaMIME.Extension()

	if mediaExt != msg.Props.Ext {
		return nil, fiber.NewError(fiber.StatusBadRequest, "the file extension detected (%s) does not match the one provided "+mediaExt)
	}

	// upload the file data in exchange for a url
	fileUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("message_medias/user-%s-%s", clientUsername, msg.ToUser), fileData, mediaExt)
	if err != nil {
		return nil, err
	}

	var msgPropsMap map[string]any

	helpers.StructToMap(msg.Props, &msgPropsMap)

	msgPropsMap["data_url"] = fileUrl

	delete(msgPropsMap, "data")

	// marshal message props to string ([]byte) for db storage
	msgProps, err := json.Marshal(msgPropsMap)
	if err != nil {
		log.Println("chatService.go: SendFile:", err)
		return nil, fiber.ErrInternalServerError
	}

	res, err := chatMessage.Send(ctx, clientUsername, msg.ToUser, msgType, string(msgProps), time.UnixMilli(msg.At).UTC())
	if err != nil {
		return nil, err
	}

	if res.PartnerRes != nil {
		go eventStreamService.Send(msg.ToUser, appTypes.ServerWSMsg{
			Event: "chat: new message",
			Data:  res.PartnerRes,
		})
	}

	return res.ClientRes, nil
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
	done, err := chatMessage.ReactTo(ctx, clientUsername, partnerUsername, msgId, reaction, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "chat: message reaction",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"reaction":         reaction,
				"at":               at,
			},
		})
	}

	return done, nil
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
