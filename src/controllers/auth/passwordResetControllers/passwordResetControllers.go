package passwordResetControllers

import (
	"context"
	"encoding/json"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/auth/passwordResetService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestPasswordReset(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body requestPasswordResetBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, sessionData, app_err := passwordResetService.RequestPasswordReset(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	sd, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("passwordResetControllers.go: RequestPasswordReset: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("passwordReset", string(sd), "/api/auth/forgot_password/confirm_action", int(time.Hour/time.Second)))

	return c.JSON(respData)
}

func ConfirmAction(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("passwordReset_sess_data").(map[string]any)

	var body confirmActionBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, newSessionData, app_err := passwordResetService.ConfirmAction(ctx, sessionData, body.Token)
	if app_err != nil {
		return app_err
	}

	nsd, err := json.Marshal(newSessionData)
	if err != nil {
		log.Println("passwordResetControllers.go: ConfirmAction: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("passwordReset", string(nsd), "/api/auth/forgot_password/reset_password", int(time.Hour/time.Second)))

	return c.JSON(respData)
}

func ResetPassword(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("passwordReset_sess_data").(map[string]any)

	var body resetPasswordBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		log.Println(val_err)
		return val_err
	}

	respData, app_err := passwordResetService.ResetPassword(ctx, sessionData, body.NewPassword)
	if app_err != nil {
		return app_err
	}

	c.ClearCookie()

	return c.JSON(respData)
}
