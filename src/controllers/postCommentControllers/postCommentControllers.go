package postCommentControllers

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/postCommentService"

	"github.com/gofiber/fiber/v2"
)

func CreateNewPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body createNewPostBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := postCommentService.CreateNewPost(ctx, clientUser.Username, body.MediaDataList, body.Type, body.Description)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
