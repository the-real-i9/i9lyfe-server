package userControllers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/userService"

	"github.com/gofiber/fiber/v2"
)

func GetClientUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetClientUser(ctx, clientUser.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func Signout(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	c.ClearCookie()

	return c.JSON(fmt.Sprintf("Bye, %s! See you again!", clientUser.Username))
}

func EditUserProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	respData, app_err := userService.EditUserProfile(ctx, clientUser.Username, body)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func ChangeUserProfilePicture(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var err error

	var body changeProfilePictureBody

	err = c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.ChangeUserProfilePicture(ctx, clientUser.Username, body.PictureData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FollowUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params followUserParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.FollowUser(ctx, clientUser.Username, params.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UnfollowUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params unfollowUserParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.UnfollowUser(ctx, clientUser.Username, params.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserMentionedPosts(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetUserMentionedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserReactedPosts(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetUserReactedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserSavedPosts(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetUserSavedPosts(ctx, clientUser.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserNotifications(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetUserNotifications(ctx, clientUser.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func ReadUserNotification(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params readUserNotificationParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.ReadUserNotification(ctx, clientUser.Username, params.NotificationId)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getUserProfileParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.GetUserProfile(ctx, clientUser.Username, params.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserFollowers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getUserFollowersParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.GetUserFollowers(ctx, clientUser.Username, params.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
func GetUserFollowing(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getUserFollowingParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.GetUserFollowing(ctx, clientUser.Username, params.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetUserPosts(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var params getUserPostsParams

	var err error

	err = c.ParamsParser(&params)
	if err != nil {
		return err
	}

	if err = params.Validate(); err != nil {
		return err
	}

	respData, app_err := userService.GetUserPosts(ctx, clientUser.Username, params.Username, c.QueryInt("limit", 20), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
