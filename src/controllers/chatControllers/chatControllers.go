package chatControllers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/chatService"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func AuthorizeUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatService.AuthorizeUpload(ctx, body.MsgType, body.MediaMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func AuthorizeVisualUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeVisualUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatService.AuthorizeVisualUpload(ctx, body.MsgType, body.MediaMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetChats(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := chatService.GetChats(ctx, clientUser.Username, helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func DeleteChat(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := chatService.DeleteChat(ctx, clientUser.Username, c.Params("partner_username"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func SendMessage(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[sendMsgAcd](actionData)

	if err := data.Validate(ctx); err != nil {
		return nil, err
	}

	return chatService.SendMessage(ctx, clientUsername, data.PartnerUsername, data.ReplyTargetMsgId, data.IsReply, helpers.ToJson(data.Msg), data.At)
}

func AckMsgDelivered(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[ackMsgDeliveredAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.AckMsgDelivered(ctx, clientUsername, data.PartnerUsername, data.MsgIdList, data.At)
}

func AckMsgRead(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[ackMsgReadAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.AckMsgRead(ctx, clientUsername, data.PartnerUsername, data.MsgIdList, data.At)
}

func GetChatHistory(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[getChatHistoryAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.GetChatHistory(ctx, clientUsername, data.PartnerUsername, helpers.CoalesceInt(data.Limit, 50), data.Cursor)
}

func ReactToMsg(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[reactToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.ReactToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.Emoji, data.At)
}

func RemoveReactionToMsg(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[removeReactionToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.RemoveReactionToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId)
}

func DeleteMsg(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {
	data := helpers.FromBtMsgPack[deleteMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.DeleteMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.DeleteFor)
}
