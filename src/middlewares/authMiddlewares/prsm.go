package authMiddlewares

import (
	"encoding/json"
	"i9lyfe/src/helpers"

	"github.com/gofiber/fiber/v2"
)

func PasswordResetSession(c *fiber.Ctx) error {
	prsStr := c.Cookies("passwordReset")

	if prsStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("out-of-turn endpoint access: complete the previous step of the password reset process")
	}

	var passwordResetSessionData map[string]any

	if err := json.Unmarshal([]byte(prsStr), &passwordResetSessionData); err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	c.Locals("passwordReset_sess_data", passwordResetSessionData)

	return c.Next()
}
