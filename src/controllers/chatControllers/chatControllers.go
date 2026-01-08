package chatControllers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/chatService"
	"i9lyfe/src/services/chatService/chatMessageService"
	"i9lyfe/src/services/uploadService/chatUploadService"

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

	respData, err := chatUploadService.Authorize(ctx, body.MsgType, body.MediaMIME)
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

	respData, err := chatUploadService.AuthorizeVisual(ctx, body.MsgType, body.MediaMIME)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetChats(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := chatService.GetChats(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func DeleteChat(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := chatService.DeleteChat(ctx, clientUser.Username, c.Params("partner_username"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func SendMessage(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[sendMsgAcd](actionData)

	if err := data.Validate(ctx); err != nil {
		return nil, err
	}

	return chatMessageService.SendMessage(ctx, clientUsername, data.PartnerUsername, data.ReplyTargetMsgId, data.IsReply, helpers.ToJson(data.Msg), data.At)
}

func AckMsgDelivered(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[ackMsgDeliveredAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatMessageService.AckMsgDelivered(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.At)
}

func AckMsgRead(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[ackMsgReadAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatMessageService.AckMsgRead(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.At)
}

func GetChatHistory(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[getChatHistoryAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	if data.Limit == 0 {
		data.Limit = 50
	}

	return chatService.GetChatHistory(ctx, clientUsername, data.PartnerUsername, data.Limit, data.Cursor)
}

func ReactToMsg(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[reactToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatMessageService.ReactToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.Emoji, data.At)
}

func RemoveReactionToMsg(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[removeReactionToMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatMessageService.RemoveReactionToMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId)
}

func DeleteMsg(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	data := helpers.MapToStruct[deleteMsgAcd](actionData)

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return chatMessageService.DeleteMsg(ctx, clientUsername, data.PartnerUsername, data.MsgId, data.DeleteFor)
}
