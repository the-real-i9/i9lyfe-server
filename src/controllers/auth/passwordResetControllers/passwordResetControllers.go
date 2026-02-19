package passwordResetControllers

import (
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/auth/passwordResetService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

// Forgot Password - Request Password Reset
//
//	@Summary		Password Reset - Step 1
//	@Description	Submit your email to request a password reset
//	@Tags			auth
//	@Produce		application/vnd.msgpack
//
//	@Param			requestPasswordResetBody	body		requestPasswordResetBody									true	"Request body"
//
//	@Success		200		{object}	passwordResetService.passReset1RespT	"Proceed to email confirmation"
//	@Header			200		{string}	Set-cookie								"Password Reset session response cookie"
//
//	@Failure		404		{object}	appErrors.HTTPError						"No user with this email exists"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/auth/forgot_password/request_password_reset [post]
func RequestPasswordReset(c fiber.Ctx) error {
	ctx := c.Context()

	var body requestPasswordResetBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, sessionData, err := passwordResetService.RequestPasswordReset(ctx, body.Email)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"passwordReset": sessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/confirm_email", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

// Forgot Password - Confirm Email
//
//	@Summary		Password Reset - Step 2
//	@Description	Provide the 6-digit token sent to email
//	@Tags			auth
//	@Produce		application/vnd.msgpack
//
//	@Param			confirmEmailBody	body		confirmEmailBody									true	"Request body"
//	@Param			Cookie	header		string									true	"Password Reset session request cookie"
//
//	@Success		200		{object}	passwordResetService.passReset2RespT	"Email confirmed. You're about to reset your password"
//	@Header			200		{string}	Set-cookie								"Password Reset session request cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError						"Incorrect or expired confirmation token"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/auth/forgot_password/confirm_email [post]
func ConfirmEmail(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(msgpack.RawMessage)

	var body confirmEmailBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, newSessionData, err := passwordResetService.ConfirmEmail(ctx, sessionData, body.Token)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"passwordReset": newSessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/reset_password", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

// Forgot Password - Reset Password
//
//	@Summary		Password Reset user - Step 3
//	@Description	Set new password
//	@Tags			auth
//	@Accept			application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			resetPasswordBody			body		resetPasswordBody									true	"Choose a new password"
//
//	@Param			Cookie				header		[]string									true	"Password Reset session request cookie"
//
//	@Success		200					{object}	passwordResetService.passReset3RespT	"Password changed successfully"
//
//	@Failure		400					{object}	appErrors.HTTPError						"Passwords mismatch."
//
//	@Failure		500					{object}	appErrors.HTTPError
//
//	@Router			/auth/forgot_password/reset_password [post]
func ResetPassword(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(msgpack.RawMessage)

	var body resetPasswordBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		helpers.LogError(err)
		return err
	}

	respData, err := passwordResetService.ResetPassword(ctx, sessionData, body.NewPassword)
	if err != nil {
		return err
	}

	c.ClearCookie()

	return c.MsgPack(respData)
}
