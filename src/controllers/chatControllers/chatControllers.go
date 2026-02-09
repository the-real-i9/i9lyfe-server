package chatControllers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/chatService"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func AuthorizeUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeUploadBody

	err := c.BodyParser(&body)
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

	return c.JSON(respData)
}

func AuthorizeVisualUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeVisualUploadBody

	err := c.BodyParser(&body)
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

	return c.JSON(respData)
}

func GetChats(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := chatService.GetChats(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func DeleteChat(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := chatService.DeleteChat(ctx, clientUser.Username, c.Params("partner_username"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func SendMessage(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[sendMsgAcd](actionData)

	if err := data.Validate(ctx); err != nil {
		return nil, err
	}

	return chatService.SendMessage(ctx, clientUsername, data.PartnerUsername, data.ReplyTargetMsgId, data.IsReply, helpers.ToJson(data.Msg), data.At)
}

func AckMsgDelivered(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[ackMsgDeliveredAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.AckMsgDelivered(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.At)
}

func AckMsgRead(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[ackMsgReadAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.AckMsgRead(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.At)
}

func GetChatHistory(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[getChatHistoryAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	if data.Limit == 0 {
		data.Limit = 50
	}

	return chatService.GetChatHistory(ctx, clientUsername, data.PartnerUsername, data.Limit, data.Cursor)
}

func ReactToMsg(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[reactToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.ReactToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.Emoji, data.At)
}

func RemoveReactionToMsg(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[removeReactionToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.RemoveReactionToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId)
}

func DeleteMsg(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {
	data := helpers.FromBtJson[deleteMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatService.DeleteMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.DeleteFor)
}
