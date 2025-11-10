package signupService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	user "i9lyfe/src/models/userModel"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/mailService"
	"i9lyfe/src/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Signup1RespT struct {
	Msg string `json:"msg"`
}

func RequestNewAccount(ctx context.Context, email string) (Signup1RespT, map[string]any, error) {
	var resp Signup1RespT

	userExists, err := user.Exists(ctx, email)
	if err != nil {
		return resp, nil, err
	}

	if userExists {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, "A user with this email already exists.")
	}

	verfCode, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Verify your email", fmt.Sprintf("<p>Your email verification code is <strong>%s</strong></p>", verfCode))

	sessionData := map[string]any{
		"email":        email,
		"vCode":        verfCode,
		"vCodeExpires": expires,
	}

	resp.Msg = fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", email)

	return resp, sessionData, nil
}

func VerifyEmail(ctx context.Context, sessionData map[string]any, inputVerfCode string) (any, map[string]any, error) {
	var sd struct {
		Email        string
		VCode        string
		VCodeExpires time.Time
	}

	helpers.ToStruct(sessionData, &sd)

	if sd.VCode != inputVerfCode {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Incorrect verification code! Check or Re-submit your email.")
	}

	if sd.VCodeExpires.Before(time.Now()) {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Verification code expired! Re-submit your email.")
	}

	go mailService.SendMail(sd.Email, "Email Verification Success", fmt.Sprintf("Your email <strong>%s</strong> has been verified!", sd.Email))

	newSessionData := map[string]any{"email": sd.Email}

	respData := map[string]any{
		"msg": fmt.Sprintf("Your email, %s, has been verified!", sd.Email),
	}

	return respData, newSessionData, nil
}

func RegisterUser(ctx context.Context, sessionData map[string]any, username, password, name, bio string, birthday int64) (any, string, error) {
	email := sessionData["email"].(string)

	userExists, err := user.Exists(ctx, username)
	if err != nil {
		return nil, "", err
	}

	if userExists {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "Username not available")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	newUser, err := user.New(ctx, email, username, hashedPassword, name, bio, birthday)
	if err != nil {
		return nil, "", err
	}

	go eventStreamService.QueueNewUserEvent(eventTypes.NewUserEvent{
		Username: newUser.Username,
		UserData: helpers.ToJson(newUser),
	})

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username:      newUser.Username,
		Name:          newUser.Name,
		ProfilePicUrl: newUser.ProfilePicUrl,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 10 days

	if err != nil {
		return nil, "", err
	}

	respData := map[string]any{
		"msg":  "Signup success!",
		"user": newUser,
	}

	return respData, authJwt, nil
}
