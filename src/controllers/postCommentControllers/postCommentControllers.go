package postCommentControllers

import (
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/postCommentService"

	"github.com/gofiber/fiber/v3"
)

func AuthorizePostUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizePostUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := postCommentService.AuthorizePostMediaUpload(ctx, body.PostType, body.MediaMIME, len(body.MediaSizes))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func AuthorizeCommentUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeCommentUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := postCommentService.AuthorizeCommAttUpload(ctx, body.AttachmentMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func CreateNewPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body createNewPostBody

	var err error

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(ctx); err != nil {
		return err
	}

	respData, err := postCommentService.CreateNewPost(ctx, clientUser.Username, body.MediaCloudNames, body.Type, body.Description, body.At)
	if err != nil {
		return err
	}

	return c.Status(201).MsgPack(respData)
}

func GetPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.GetPost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func DeletePost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.DeletePost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func ReactToPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body reactToPostBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := postCommentService.ReactToPost(ctx, clientUser.Username, c.Params("postId"), body.Emoji, body.At)
	if err != nil {
		return err
	}

	return c.Status(201).MsgPack(respData)
}

func GetReactorsToPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := postCommentService.GetReactorsToPost(ctx, clientUser.Username, c.Params("postId"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

/* func GetReactorsWithReactionToPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	reaction, err := url.PathUnescape(c.Params("reaction"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}

	respData, err := postCommentService.GetReactorsWithReactionToPost(ctx, clientUser.Username, c.Params("postId"), reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
} */

func RemoveReactionToPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.RemoveReactionToPost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func CommentOnPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body commentOnPostBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(ctx); err != nil {
		return err
	}

	respData, err := postCommentService.CommentOnPost(ctx, clientUser.Username, c.Params("postId"), body.CommentText, body.AttachmentCloudName, body.At)
	if err != nil {
		return err
	}

	return c.Status(201).MsgPack(respData)
}

func GetCommentsOnPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := postCommentService.GetCommentsOnPost(ctx, clientUser.Username, c.Params("postId"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.GetComment(ctx, clientUser.Username, c.Params("commentId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func RemoveCommentOnPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.RemoveCommentOnPost(ctx, clientUser.Username, c.Params("postId"), c.Params("commentId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func ReactToComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body reactToCommentBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := postCommentService.ReactToComment(ctx, clientUser.Username, c.Params("commentId"), body.Emoji, body.At)
	if err != nil {
		return err
	}

	return c.Status(201).MsgPack(respData)
}

func GetReactorsToComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := postCommentService.GetReactorsToComment(ctx, clientUser.Username, c.Params("commentId"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

/* func GetReactorsWithReactionToComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	reaction, err := url.PathUnescape(c.Params("reaction"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}

	respData, err := postCommentService.GetReactorsWithReactionToComment(ctx, clientUser.Username, c.Params("commentId"), reaction, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
} */

func RemoveReactionToComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.RemoveReactionToComment(ctx, clientUser.Username, c.Params("commentId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func CommentOnComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body commentOnCommentBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(ctx); err != nil {
		return err
	}

	respData, err := postCommentService.CommentOnComment(ctx, clientUser.Username, c.Params("commentId"), body.CommentText, body.AttachmentCloudName, body.At)
	if err != nil {
		return err
	}

	return c.Status(201).MsgPack(respData)
}

func GetCommentsOnComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := postCommentService.GetCommentsOnComment(ctx, clientUser.Username, c.Params("commentId"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func RemoveCommentOnComment(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.RemoveCommentOnComment(ctx, clientUser.Username, c.Params("parentCommentId"), c.Params("childCommentId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func RepostPost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.RepostPost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func SavePost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.SavePost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func UnsavePost(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := postCommentService.UnsavePost(ctx, clientUser.Username, c.Params("postId"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}
