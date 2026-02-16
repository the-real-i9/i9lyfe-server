package signinService

import (
	"context"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/securityServices"
	"i9lyfe/src/services/userService"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

type signinRespT struct {
	Msg  string             `msgpack:"msg"`
	User UITypes.ClientUser `msgpack:"user"`
}

func Signin(ctx context.Context, emailOrUsername, inputPassword string) (signinRespT, string, error) {
	var resp signinRespT

	theUser, err := userService.SigninUserFind(ctx, emailOrUsername)
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
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))

	if err != nil {
		return resp, "", err
	}

	theUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(theUser.ProfilePicUrl)

	resp.Msg = "Signin success!"
	resp.User = UITypes.ClientUser{Username: theUser.Username, Name: theUser.Name, ProfilePicUrl: theUser.ProfilePicUrl}

	return resp, authJwt, nil
}
