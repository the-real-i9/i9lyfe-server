package publicRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthOptional)
}
