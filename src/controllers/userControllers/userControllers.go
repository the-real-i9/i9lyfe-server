package userControllers

import (
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/userService"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Get session user
//
//	@Summary		Get session user
//	@Description	Get info on the user currently in session
//	@Tags			app/private
//	@Produce		application/vnd.msgpack
//
//	@Param			Cookie	header		[]string			true	"User session request cookie"
//
//	@Success		200		{object}	UITypes.ClientUser	"User info"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/app/private/me [get]
func GetSessionUser(c fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	user, err := userService.SigninUserFind(c.Context(), clientUser.Username)
	if err != nil {
		return err
	}

	return c.MsgPack(UITypes.ClientUser{Username: user.Username, Name: user.Name, ProfilePicUrl: user.ProfilePicUrl})
}

// Signout session user
//
//	@Summary		Signout user
//	@Description	Signout the user currently in session
//	@Tags			app/private
//	@Produce		application/vnd.msgpack
//
//	@Param			Cookie	header		[]string	true	"User session request cookie"
//
//	@Success		200		{string}	string		"Bye message"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/app/private/me/signout [get]
func Signout(c fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	c.ClearCookie()

	return c.MsgPack(fmt.Sprintf("Bye, %s! See you again!", clientUser.Username))
}

// Edit user profile
//
//	@Summary		Edit user profile
//	@Description	Edit user profile
//	@Tags			app/private
//	@Accepts		application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			name		body		string		false	"User's Name field"
//	@Param			birthday	body		int			false	"User's Birthday field in milliseconds since Unix Epoch"
//	@Param			bio			body		string		false	"User's Bio field"
//
//	@Param			Cookie		header		[]string	true	"User session request cookie"
//
//	@Success		200			{object}	boolean		"Done"
//
//	@Failure		500			{object}	appErrors.HTTPError
//
//	@Router			/app/private/me/edit_profile [put]
func EditUserProfile(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body editProfileBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := userService.EditUserProfile(ctx, clientUser.Username, body)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func AuthorizePPicUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizePPicUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := userService.AuthorizePPicUpload(ctx, body.PicMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

// Change user profile picture
//
//	@Summary		Change user profile picture
//	@Description	Change user profile picture
//	@Tags			app/private
//	@Accepts		application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			picture_data	body		[]byte		true	"Profile picture data"
//
//	@Param			Cookie			header		[]string	true	"User session request cookie"
//
//	@Success		200				{object}	boolean		"Done"
//
//	@Failure		500				{object}	appErrors.HTTPError
//
//	@Router			/app/private/me/change_profile_picture [put]
func ChangeUserProfilePicture(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body changeProfilePictureBody

	err = c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(ctx); err != nil {
		return err
	}

	respData, err := userService.ChangeUserProfilePicture(ctx, clientUser.Username, body.ProfilePicCloudName)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

// Follow user
//
//	@Summary		Follow user
//	@Description	Follow user
//	@Tags			app/private
//	@Produce		application/vnd.msgpack
//
//	@Param			username	path		string				true	"User to follow"
//
//	@Param			Cookie		header		[]string			true	"User session request cookie"
//
//	@Success		200			{object}	boolean				"Done"
//
//	@Success		400			{object}	appErrors.HTTPError	"Validation error | User trying to follow self"
//
//	@Failure		500			{object}	appErrors.HTTPError
//
//	@Router			/app/private/users/:username/follow [post]
func FollowUser(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.FollowUser(ctx, clientUser.Username, c.Params("username"), time.Now().UnixMilli())
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

// Unfollow user
//
//	@Summary		Unfollow user
//	@Description	Unfollow user
//	@Tags			app/private
//	@Produce		application/vnd.msgpack
//
//	@Param			username	path		string		true	"User to unfollow"
//
//	@Param			Cookie		header		[]string	true	"User session request cookie"
//
//	@Success		200			{object}	boolean		"Done"
//
//	@Failure		500			{object}	appErrors.HTTPError
//
//	@Router			/app/private/users/:username/unfollow [delete]
func UnfollowUser(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.UnfollowUser(ctx, clientUser.Username, c.Params("username"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserMentionedPosts(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserMentionedPosts(ctx, clientUser.Username, helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserReactedPosts(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserReactedPosts(ctx, clientUser.Username, helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserSavedPosts(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserSavedPosts(ctx, clientUser.Username, helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserNotifications(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
		Year   int64
		Month  int64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserNotifications(ctx, clientUser.Username, helpers.CoalesceInt(query.Year, int64(time.Now().Year())), helpers.CoalesceInt(query.Month, int64(time.Now().Month())), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func ReadUserNotification(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.ReadUserNotification(ctx, clientUser.Username, c.Params("year", fmt.Sprint(time.Now().Year())), c.Params("month", fmt.Sprintf("%d", time.Now().Month())), c.Params("notification_id"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserProfile(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserProfile(ctx, clientUser.Username, c.Params("username"))
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserFollowers(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserFollowers(ctx, clientUser.Username, c.Params("username"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}
func GetUserFollowings(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserFollowings(ctx, clientUser.Username, c.Params("username"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetUserPosts(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetUserPosts(ctx, clientUser.Username, c.Params("username"), helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}
