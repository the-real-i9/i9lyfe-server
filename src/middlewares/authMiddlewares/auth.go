package authMiddlewares

import (
	"encoding/json"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/securityServices"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/zishang520/socket.io/v2/socket"
)

func UserAuthRequired(c *fiber.Ctx) error {
	usStr := c.Cookies("user")

	if usStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	var userSessionData map[string]string

	if err := json.Unmarshal([]byte(usStr), &userSessionData); err != nil {
		log.Println("auth.go: UserAuth: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}

func UserAuthOptional(c *fiber.Ctx) error {
	usStr := c.Cookies("user")

	if usStr == "" {
		c.Locals("user", nil)
		return c.Next()
	}

	var userSessionData map[string]string

	if err := json.Unmarshal([]byte(usStr), &userSessionData); err != nil {
		log.Println("auth.go: UserAuth: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}

func UserAuthSocket(cliSocket *socket.Socket, next func(*socket.ExtendedError)) {
	// authenticate user
	usStr, found := cliSocket.Request().Headers().Get("user")
	if !found {
		next(socket.NewExtendedError("authentication required", nil))
		return
	}

	var userSessionData map[string]string

	if err := json.Unmarshal([]byte(usStr), &userSessionData); err != nil {
		log.Println("socket.go: InitSocket: json.Unmarshal:", err)
		next(socket.NewExtendedError(fiber.ErrInternalServerError.Message, nil))
		return
	}

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		f_err := err.(*fiber.Error)
		next(socket.NewExtendedError(f_err.Message, nil))
		return
	}

	cliSocket.SetData(clientUser)

	next(nil)
}
