package signinService

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/models/userModel"
	"i9lyfe/src/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type SigninRespT struct {
	Msg  string                `json:"msg"`
	User userModel.ToAuthUserT `json:"user"`
}

func Signin(ctx context.Context, emailOrUsername, inputPassword string) (SigninRespT, string, error) {
	var resp SigninRespT

	theUser, err := userModel.AuthFind(ctx, emailOrUsername)
	if err != nil {
		return resp, "", err
	}

	if theUser == nil {
		return resp, "", fiber.NewError(fiber.StatusNotFound, "Incorrect email or password")
	}

	hashedPassword := theUser.Password

	yes, err := securityServices.PasswordMatchesHash(hashedPassword, inputPassword)
	if err != nil {
		return resp, "", err
	}

	if !yes {
		return resp, "", fiber.NewError(fiber.StatusNotFound, "Incorrect email or password")
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username:      theUser.Username,
		Name:          theUser.Name,
		ProfilePicUrl: theUser.ProfilePicUrl,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))

	if err != nil {
		return resp, "", err
	}

	resp.Msg = "Signin success!"
	resp.User = *theUser

	return resp, authJwt, nil
}
