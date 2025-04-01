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

	var err error

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CreateNewPost(ctx, clientUser.Username, body.MediaDataList, body.Type, body.Description)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getPostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetPost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func DeletePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getPostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetPost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func ReactToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params reactToPostParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	var body reactToPostBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.ReactToPost(ctx, clientUser.Username, params.PostId, body.Reaction)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
