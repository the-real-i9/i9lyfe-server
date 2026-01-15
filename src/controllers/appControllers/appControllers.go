package appControllers

/* import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/appService"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FetchPosts(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	typesFilter := strings.Split(c.Query("types"), ",")
	hashtagFilter := strings.Split(c.Query("hashtags"), ",")

	respData, err := appService.FetchPosts(ctx, clientUser.Username, typesFilter, hashtagFilter)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func FetchUsers(c *fiber.Ctx) error {
	ctx := c.Context()

	respData, err := appService.FetchUsers(ctx)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func FetchHashtags(c *fiber.Ctx) error {
	ctx := c.Context()

	respData, err := appService.FetchHashtags(ctx)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func Search(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	term := c.Query("term")
	filter := c.Query("filter")

	respData, err := appService.Search(ctx, clientUser.Username, term, filter, c.QueryInt("limit", 50), int64(c.QueryInt("offset")))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
*/
