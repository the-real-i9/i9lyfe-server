package privateRoutes

import (
	"i9lyfe/src/middlewares/authMiddlewares"
	"i9lyfe/src/routes/privateRoutes/chatRoutes"
	"i9lyfe/src/routes/privateRoutes/postCommentRoutes"
	"i9lyfe/src/routes/privateRoutes/userPrivateRoutes"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Use(authMiddlewares.UserAuthRequired)

	router.Use(postCommentRoutes.Init)
	router.Use(chatRoutes.Init)
	router.Use(userPrivateRoutes.Init)
}
