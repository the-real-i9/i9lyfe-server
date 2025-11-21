package signupControllers

import (
	"encoding/json"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/auth/signupService"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Signup - Request New Account
//
//	@Summary		Signup user - Step 1
//	@Description	Submit email to request a new account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			email	body		string						true	"Provide your email address"
//
//	@Success		200		{object}	signupService.signup1RespT	"Proceed to email verification"
//	@Header			200		{array}		Set-Cookie					"Signup session response cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError			"An account with email already exists"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/auth/signup/request_new_account [post]
func RequestNewAccount(c *fiber.Ctx) error {
	ctx := c.Context()

	var body requestNewAccountBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, sessionData, app_err := signupService.RequestNewAccount(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	sd, err := json.Marshal(sessionData)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("signup", string(sd), "/api/auth/signup/verify_email", int(time.Hour/time.Second)))

	return c.JSON(respData)
}

// Signup - Verify Email
//
//	@Summary		Signup user - Step 2
//	@Description	Provide the 6-digit code sent to email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			code	body		string						true	"6-digit code"
//	@Param			Cookie	header		[]string						true	"Signup session request cookie"
//
//	@Success		200		{object}	signupService.signup2RespT	"Email verified"
//	@Header			200		{array}		Set-Cookie					"Signup session response cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError			"Incorrect or expired verification code"
//	@Header			400		{array}		Set-Cookie					"Signup session response cookie"
//
//	@Failure		500		{object}	appErrors.HTTPError
//	@Header			500		{array}		Set-Cookie	"Signup session response cookie"
//
//	@Router			/auth/signup/verify_email [post]
func VerifyEmail(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(map[string]any)

	var body verifyEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, newSessionData, app_err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if app_err != nil {
		return app_err
	}

	nsd, err := json.Marshal(newSessionData)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("signup", string(nsd), "/api/auth/signup/register_user", int(time.Hour/time.Second)))

	return c.JSON(respData)
}

// Signup - Register user
//
//	@Summary		Signup user - Step 3
//	@Description	Provide remaining user credentials
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			username	body		string						true	"Choose a username"
//	@Param			password	body		string						true	"Choose a password"
//	@Param			name		body		string						true	"User display name"
//	@Param			birthday	body		int						true	"User birthday in milliseconds since Unix Epoch"
//	@Param			bio			body		string						false	"User bio (optional)"
//
//	@Param			Cookie		header		[]string						true	"Signup session request cookie"
//
//	@Success		200			{object}	signupService.signup3RespT	"Signup Success"
//	@Header			200			{array}		Set-Cookie					"User session response cookie containing auth JWT"
//
//	@Failure		400			{object}	appErrors.HTTPError			"Incorrect or expired verification code"
//	@Header			400			{array}		Set-Cookie					"Signup session response cookie"
//
//	@Failure		500			{object}	appErrors.HTTPError
//	@Header			500			{array}		Set-Cookie	"Signup session response cookie"
//
//	@Router			/auth/signup/register_user [post]
func RegisterUser(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(map[string]any)

	var body registerUserBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		helpers.LogError(val_err)
		return val_err
	}

	respData, authJwt, app_err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password, body.Name, body.Bio, body.Birthday)
	if app_err != nil {
		return app_err
	}

	usd, err := json.Marshal(map[string]any{"authJwt": authJwt})
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("user", string(usd), "/api/app", int(10*24*time.Hour/time.Second)))

	return c.Status(fiber.StatusCreated).JSON(respData)
}
