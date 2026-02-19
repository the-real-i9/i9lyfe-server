package passwordResetService

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/mailService"
	"i9lyfe/src/services/securityServices"
	"i9lyfe/src/services/userService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type passReset1RespT struct {
	Msg string `msgpack:"msg" json:"msg"`
}

func RequestPasswordReset(ctx context.Context, email string) (passReset1RespT, map[string]any, error) {
	var resp passReset1RespT

	userExists, err := userService.UserExists(ctx, email)
	if err != nil {
		return resp, nil, err
	}

	if !userExists {
		return resp, nil, fiber.NewError(fiber.StatusNotFound, "No user with this email exists.")
	}

	pwdrToken, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Confirm your email: Password Reset", fmt.Sprintf("<p>Your password reset token is <strong>%s</strong>.</p>", pwdrToken))

	sessionData := map[string]any{
		"email":            email,
		"pwdrToken":        pwdrToken,
		"pwdrTokenExpires": expires,
	}

	resp.Msg = fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", email)

	return resp, sessionData, nil
}

type passReset2RespT struct {
	Msg string `msgpack:"msg" json:"msg"`
}

func ConfirmEmail(ctx context.Context, sessionData msgpack.RawMessage, inputResetToken string) (passReset2RespT, map[string]any, error) {
	var resp passReset2RespT

	sd := helpers.FromBtMsgPack[struct {
		Email            string    `msgpack:"email"`
		PwdrToken        string    `msgpack:"pwdrToken"`
		PwdrTokenExpires time.Time `msgpack:"pwdrTokenExpires"`
	}](sessionData)

	if sd.PwdrToken != inputResetToken {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, "Incorrect password reset token! Check or Re-submit your email.")
	}

	if sd.PwdrTokenExpires.Before(time.Now()) {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, "Password reset token expired! Re-submit your email.")
	}

	newSessionData := map[string]any{"email": sd.Email}

	resp.Msg = fmt.Sprintf("%s, you're about to reset your password!", sd.Email)

	return resp, newSessionData, nil
}

type passReset3RespT struct {
	Msg string `msgpack:"msg" json:"msg"`
}

func ResetPassword(ctx context.Context, sessionData msgpack.RawMessage, newPassword string) (passReset3RespT, error) {
	var resp passReset3RespT

	email := helpers.FromBtMsgPack[struct {
		Email string `msgpack:"email"`
	}](sessionData).Email

	hashedPassword, err := securityServices.HashPassword(newPassword)
	if err != nil {
		return resp, err
	}

	done, err := userService.ChangeUserPassword(ctx, email, hashedPassword)
	if err != nil {
		return resp, err
	}

	if done {
		go mailService.SendMail(email, "Password Reset Success", fmt.Sprintf("<p>%s, your password has been changed successfully!</p>", email))

		resp.Msg = "Your password has been changed successfully"
	}

	return resp, nil
}
