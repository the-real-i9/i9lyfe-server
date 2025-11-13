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

	respData, app_err := postCommentService.CreateNewPost(ctx, clientUser, body.MediaDataList, body.Type, body.Description, body.At)
	if app_err != nil {
		return app_err
	}

	return c.Status(201).JSON(respData)
}

func GetPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetPost(ctx, clientUser.Username, c.Params("postId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func DeletePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.DeletePost(ctx, clientUser.Username, c.Params("postId"))
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

	var body reactToPostBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.ReactToPost(ctx, clientUser, c.Params("postId"), body.Emoji, body.At)
	if app_err != nil {
		return app_err
	}

	return c.Status(201).JSON(respData)
}

func GetReactorsToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetReactorsToPost(ctx, clientUser.Username, c.Params("postId"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

/* func GetReactorsWithReactionToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	reaction, err := url.PathUnescape(c.Params("reaction"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}

	respData, app_err := postCommentService.GetReactorsWithReactionToPost(ctx, clientUser.Username, c.Params("postId"), reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
} */

func RemoveReactionToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.RemoveReactionToPost(ctx, clientUser.Username, c.Params("postId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func CommentOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body commentOnPostBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CommentOnPost(ctx, clientUser, c.Params("postId"), body.CommentText, body.AttachmentData, body.At)
	if app_err != nil {
		return app_err
	}

	return c.Status(201).JSON(respData)
}

func GetCommentsOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetCommentsOnPost(ctx, clientUser.Username, c.Params("postId"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetComment(ctx, clientUser.Username, c.Params("commentId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func RemoveCommentOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.RemoveCommentOnPost(ctx, clientUser.Username, c.Params("postId"), c.Params("commentId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func ReactToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body reactToCommentBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.ReactToComment(ctx, clientUser, c.Params("commentId"), body.Emoji, body.At)
	if app_err != nil {
		return app_err
	}

	return c.Status(201).JSON(respData)
}

func GetReactorsToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetReactorsToComment(ctx, clientUser.Username, c.Params("commentId"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

/* func GetReactorsWithReactionToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	reaction, err := url.PathUnescape(c.Params("reaction"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}

	respData, app_err := postCommentService.GetReactorsWithReactionToComment(ctx, clientUser.Username, c.Params("commentId"), reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
} */

func RemoveReactionToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.RemoveReactionToComment(ctx, clientUser.Username, c.Params("commentId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func CommentOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body commentOnCommentBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CommentOnComment(ctx, clientUser, c.Params("commentId"), body.CommentText, body.AttachmentData, body.At)
	if app_err != nil {
		return app_err
	}

	return c.Status(201).JSON(respData)
}

func GetCommentsOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.GetCommentsOnComment(ctx, clientUser.Username, c.Params("commentId"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func RemoveCommentOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.RemoveCommentOnComment(ctx, clientUser, c.Params("parentCommentId"), c.Params("childCommentId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func RepostPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.RepostPost(ctx, clientUser, c.Params("postId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func SavePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.SavePost(ctx, clientUser.Username, c.Params("postId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UnsavePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := postCommentService.UnsavePost(ctx, clientUser.Username, c.Params("postId"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
