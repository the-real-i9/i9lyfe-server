package signupService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"

	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/mailService"
	"i9lyfe/src/services/securityServices"
	"i9lyfe/src/services/userService"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type signup1RespT struct {
	Msg string `msgpack:"msg"`
}

func RequestNewAccount(ctx context.Context, email string) (signup1RespT, map[string]any, error) {
	var resp signup1RespT

	userExists, err := userService.UserExists(ctx, email)
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

type signup2RespT struct {
	Msg string `msgpack:"msg"`
}

func VerifyEmail(ctx context.Context, sessionData msgpack.RawMessage, inputVerfCode string) (signup2RespT, map[string]any, error) {
	var resp signup2RespT

	sd := helpers.FromBtMsgPack[struct {
		Email        string    `msgpack:"email"`
		VCode        string    `msgpack:"vCode"`
		VCodeExpires time.Time `msgpack:"vCodeExpires"`
	}](sessionData)

	if sd.VCode != inputVerfCode {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, "Incorrect verification code! Check or Re-submit your email.")
	}

	if sd.VCodeExpires.Before(time.Now()) {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, "Verification code expired! Re-submit your email.")
	}

	go mailService.SendMail(sd.Email, "Email Verification Success", fmt.Sprintf("Your email <strong>%s</strong> has been verified!", sd.Email))

	newSessionData := map[string]any{"email": sd.Email}

	resp.Msg = fmt.Sprintf("Your email, %s, has been verified!", sd.Email)

	return resp, newSessionData, nil
}

type signup3RespT struct {
	Msg  string             `msgpack:"msg"`
	User UITypes.ClientUser `msgpack:"user"`
}

func RegisterUser(ctx context.Context, sessionData msgpack.RawMessage, username, name, bio string, birthday int64, password string) (signup3RespT, string, error) {
	var resp signup3RespT

	email := helpers.FromBtMsgPack[struct {
		Email string `msgpack:"email"`
	}](sessionData).Email

	userExists, err := userService.UserExists(ctx, username)
	if err != nil {
		return resp, "", err
	}

	if userExists {
		return resp, "", fiber.NewError(fiber.StatusBadRequest, "Username not available")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return resp, "", err
	}

	newUser, err := userService.NewUser(ctx, email, username, name, bio, birthday, hashedPassword)
	if err != nil {
		return resp, "", err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 10 days

	if err != nil {
		return resp, "", err
	}

	newUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(newUser.ProfilePicUrl)

	resp.Msg = "Signup success!"
	resp.User = UITypes.ClientUser{Username: newUser.Username, Name: newUser.Name, ProfilePicUrl: newUser.ProfilePicUrl}

	return resp, authJwt, nil
}
