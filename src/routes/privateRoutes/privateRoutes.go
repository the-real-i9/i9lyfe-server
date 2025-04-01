package privateRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/routes/privateRoutes/chatRoute"
	"i9lyfe/src/routes/privateRoutes/postCommentRoute"
	"i9lyfe/src/routes/privateRoutes/privateUserRoute"
	"i9lyfe/src/routes/privateRoutes/wsRoute"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthRequired)

	router.Use(postCommentRoute.Route)
	router.Use(chatRoute.Route)
	router.Use(privateUserRoute.Route)
	router.Use(wsRoute.Route)
}
