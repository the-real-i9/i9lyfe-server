package appControllers

/* import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/appService"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FetchPosts(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	typesFilter := strings.Split(c.Query("types"), ",")
	hashtagFilter := strings.Split(c.Query("hashtags"), ",")

	respData, app_err := appService.FetchPosts(ctx, clientUser.Username, typesFilter, hashtagFilter)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FetchUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	respData, app_err := appService.FetchUsers(ctx)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FetchHashtags(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	respData, app_err := appService.FetchHashtags(ctx)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func Search(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	term := c.Query("term")
	filter := c.Query("filter")

	respData, app_err := appService.Search(ctx, clientUser.Username, term, filter, c.QueryInt("limit", 50), int64(c.QueryInt("offset")))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
*/
