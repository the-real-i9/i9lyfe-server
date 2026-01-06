package chatControllers

import (
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/chatService"
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
