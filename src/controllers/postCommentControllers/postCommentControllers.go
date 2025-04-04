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

	respData, app_err := postCommentService.DeletePost(ctx, clientUser.Username, params.PostId)
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

func GetReactorsToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getReactorsToPostParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetReactorsToPost(ctx, clientUser.Username, params.PostId, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetReactorsWithReactionToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getReactorsWithReactionToPostParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetReactorsWithReactionToPost(ctx, clientUser.Username, params.PostId, params.Reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UndoReactionToPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params undoReactionToPostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.UndoReactionToPost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func CommentOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params commentOnPostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	var body commentOnPostBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CommentOnPost(ctx, clientUser.Username, params.PostId, body.CommentText, body.AttachmentData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetCommentsOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getCommentsOnPostParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetCommentsOnPost(ctx, clientUser.Username, params.PostId, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getCommentParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetComment(ctx, clientUser.Username, params.CommentId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func RemoveCommentOnPost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params removeCommentOnPostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.RemoveCommentOnPost(ctx, clientUser.Username, params.PostId, params.CommentId)
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

	var params reactToCommentParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	var body reactToCommentBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.ReactToComment(ctx, clientUser.Username, params.CommentId, body.Reaction)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetReactorsToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getReactorsToCommentParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetReactorsToComment(ctx, clientUser.Username, params.CommentId, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetReactorsWithReactionToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getReactorsWithReactionToCommentParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetReactorsWithReactionToComment(ctx, clientUser.Username, params.CommentId, params.Reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UndoReactionToComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params undoReactionToCommentParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.UndoReactionToComment(ctx, clientUser.Username, params.CommentId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func CommentOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params commentOnCommentParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	var body commentOnCommentBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CommentOnComment(ctx, clientUser.Username, params.CommentId, body.CommentText, body.AttachmentData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetCommentsOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var params getCommentsOnCommentParams

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.GetCommentsOnComment(ctx, clientUser.Username, params.CommentId, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func RemoveCommentOnComment(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params removeCommentOnCommentParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.RemoveCommentOnComment(ctx, clientUser.Username, params.ParentCommentId, params.ChildCommentId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func CreateRepost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params createRepostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.CreateRepost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func SavePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params savePostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.SavePost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UndoSavePost(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params undoSavePostParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := postCommentService.UndoSavePost(ctx, clientUser.Username, params.PostId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
