package privateRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/routes/privateRoutes/chatRoutes"
	"i9lyfe/src/routes/privateRoutes/postCommentRoutes"
	"i9lyfe/src/routes/privateRoutes/userPrivateRoutes"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthRequired)

	router.Use(postCommentRoutes.Routes)
	router.Use(chatRoutes.Routes)
	router.Use(userPrivateRoutes.Routes)
}
