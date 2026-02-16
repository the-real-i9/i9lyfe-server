package privateRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/routes/privateRoutes/chatRoute"
	"i9lyfe/src/routes/privateRoutes/postCommentRoute"
	"i9lyfe/src/routes/privateRoutes/privateUserRoute"
	"i9lyfe/src/routes/privateRoutes/wsRoute"

	"github.com/gofiber/fiber/v3"
)

func Routes(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthRequired)

	router.Route("", postCommentRoute.Route)
	router.Route("", privateUserRoute.Route)
	router.Route("", chatRoute.Route)
	router.Route("", wsRoute.Route)
}
