package testSessionRoute

import (
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Post("/signup/request_new_account", func(c fiber.Ctx) error {
		var body struct {
			Email string `msgpack:"email"`
		}

		if err := c.Bind().MsgPack(&body); err != nil {
			return err
		}

		verfCode, expires := securityServices.GenerateTokenCodeExp()

		reqSession := map[string]any{
			"signup": map[string]any{
				"email":        body.Email,
				"vCode":        verfCode,
				"vCodeExpires": expires,
			},
		}

		c.Cookie(helpers.Session(reqSession, "/api/auth/signup/verify_email", int(time.Hour/time.Second)))

		return c.SendStatus(200)
	})

	router.Post("/signup/verify_email", func(c fiber.Ctx) error {
		var body struct {
			Email string `msgpack:"email"`
		}

		if err := c.Bind().MsgPack(&body); err != nil {
			return err
		}

		reqSession := map[string]any{
			"signup": map[string]any{"email": body.Email},
		}

		c.Cookie(helpers.Session(reqSession, "/api/auth/signup/register_user", int(time.Hour/time.Second)))

		return c.SendStatus(200)
	})

	router.Post("/forgot_password/request_password_reset", func(c fiber.Ctx) error {
		var body struct {
			Email string `msgpack:"email"`
		}

		if err := c.Bind().MsgPack(&body); err != nil {
			return err
		}

		pwdrToken, expires := securityServices.GenerateTokenCodeExp()

		reqSession := map[string]any{
			"passwordReset": map[string]any{
				"email":            body.Email,
				"pwdrToken":        pwdrToken,
				"pwdrTokenExpires": expires,
			},
		}

		c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/confirm_email", int(time.Hour/time.Second)))

		return c.SendStatus(200)
	})

	router.Post("/forgot_password/confirm_email", func(c fiber.Ctx) error {
		var body struct {
			Email string `msgpack:"email"`
		}

		if err := c.Bind().MsgPack(&body); err != nil {
			return err
		}

		reqSession := map[string]any{
			"passwordReset": map[string]any{"email": body.Email},
		}

		c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/reset_password", int(time.Hour/time.Second)))

		return c.SendStatus(200)
	})

	router.Post("auth_user", func(c fiber.Ctx) error {
		var body struct {
			Username string `msgpack:"username"`
		}

		if err := c.Bind().MsgPack(&body); err != nil {
			return err
		}

		authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
			Username: body.Username,
		}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 10 days
		if err != nil {
			return err
		}

		reqSession := map[string]any{
			"user": map[string]any{"authJwt": authJwt},
		}

		c.Cookie(helpers.Session(reqSession, "/api/app", int(10*24*time.Hour/time.Second)))

		return c.SendStatus(200)
	})
}
