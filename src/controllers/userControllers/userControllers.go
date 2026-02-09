package userControllers

import (
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/userService"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Get session user
//
//	@Summary		Get session user
//	@Description	Get info on the user currently in session
//	@Tags			app/private
//	@Produce		json
//
//	@Param			Cookie	header		[]string			true	"User session request cookie"
//
//	@Success		200		{object}	UITypes.ClientUser	"User info"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/app/private/me [get]
func GetSessionUser(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	user, err := userService.SigninUserFind(c.Context(), clientUser.Username)
	if err != nil {
		return err
	}

	user.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(user.ProfilePicUrl)

	return c.JSON(UITypes.ClientUser{Username: user.Username, Name: user.Name, ProfilePicUrl: user.ProfilePicUrl})
}

// Signout session user
//
//	@Summary		Signout user
//	@Description	Signout the user currently in session
//	@Tags			app/private
//	@Produce		json
//
//	@Param			Cookie	header		[]string	true	"User session request cookie"
//
//	@Success		200		{string}	string		"Bye message"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/app/private/me/signout [get]
func Signout(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	c.ClearCookie()

	return c.JSON(fmt.Sprintf("Bye, %s! See you again!", clientUser.Username))
}

// Edit user profile
//
//	@Summary		Edit user profile
//	@Description	Edit user profile
//	@Tags			app/private
//	@Accepts		json
//	@Produce		json
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
func EditUserProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body editProfileBody

	err = c.BodyParser(&body)
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

	return c.JSON(respData)
}

func AuthorizePPicUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizePPicUploadBody

	err := c.BodyParser(&body)
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

	return c.JSON(respData)
}

// Change user profile picture
//
//	@Summary		Change user profile picture
//	@Description	Change user profile picture
//	@Tags			app/private
//	@Accepts		json
//	@Produce		json
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
func ChangeUserProfilePicture(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body changeProfilePictureBody

	err = c.BodyParser(&body)
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

	return c.JSON(respData)
}

// Follow user
//
//	@Summary		Follow user
//	@Description	Follow user
//	@Tags			app/private
//	@Produce		json
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
func FollowUser(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.FollowUser(ctx, clientUser.Username, c.Params("username"), time.Now().UnixMilli())
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

// Unfollow user
//
//	@Summary		Unfollow user
//	@Description	Unfollow user
//	@Tags			app/private
//	@Produce		json
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
func UnfollowUser(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.UnfollowUser(ctx, clientUser.Username, c.Params("username"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserMentionedPosts(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserMentionedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserReactedPosts(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserReactedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserSavedPosts(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserSavedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserNotifications(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserNotifications(ctx, clientUser.Username, c.QueryInt("year", time.Now().Year()), c.Query("month", time.Now().Month().String()), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func ReadUserNotification(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.ReadUserNotification(ctx, clientUser.Username, c.Params("year", fmt.Sprint(time.Now().Year())), c.Params("month", time.Now().Month().String()), c.Params("notification_id"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserProfile(ctx, clientUser.Username, c.Params("username"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserFollowers(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserFollowers(ctx, clientUser.Username, c.Params("username"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
func GetUserFollowings(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserFollowings(ctx, clientUser.Username, c.Params("username"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetUserPosts(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetUserPosts(ctx, clientUser.Username, c.Params("username"), c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
