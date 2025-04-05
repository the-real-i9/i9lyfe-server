package publicRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/routes/publicRoutes/appRoute"
	"i9lyfe/src/routes/publicRoutes/publicUserRoute"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthOptional)

	router.Use(publicUserRoute.Route)
	router.Use(appRoute.Route)
}
