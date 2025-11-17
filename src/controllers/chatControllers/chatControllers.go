package chatControllers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/chatService"

	"github.com/gofiber/fiber/v2"
)

func GetChats(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := chatService.GetChats(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func DeleteChat(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := chatService.DeleteChat(ctx, clientUser.Username, c.Params("partner_username"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
