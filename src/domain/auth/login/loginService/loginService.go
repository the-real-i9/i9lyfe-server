package loginService

import (
	"context"
	"i9lyfe/src/domain/user/userService"
	"i9lyfe/src/services/securityServices"
	"i9lyfe/src/types/UITypes"
	"i9lyfe/src/types/appTypes"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

type loginRespT struct {
	Msg  string             `msgpack:"msg"`
	User UITypes.ClientUser `msgpack:"user"`
}

func Login(ctx context.Context, emailOrUsername, inputPassword string) (loginRespT, string, error) {
	var resp loginRespT

	theUser, err := userService.LoginUserFind(ctx, emailOrUsername)
	if err != nil {
		return resp, "", err
	}

	if theUser.Username == "" {
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

	resp.Msg = "Login success!"
	resp.User = UITypes.ClientUser{Username: theUser.Username, Name: theUser.Name, ProfilePicUrl: theUser.ProfilePicUrl}

	return resp, authJwt, nil
}
