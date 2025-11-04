package signinControllers

import (
	"context"
	"encoding/json"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/auth/signinService"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body signInBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, authJwt, app_err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)
	if app_err != nil {
		return app_err
	}

	usd, err := json.Marshal(map[string]any{"authJwt": authJwt})
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("user", string(usd), "/api/app", int(10*24*time.Hour/time.Second)))

	return c.JSON(respData)
}
