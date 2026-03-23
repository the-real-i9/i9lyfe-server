package chatRoutes

import (
	CC "i9lyfe/src/domain/chat/chatControllers"

	"github.com/gofiber/fiber/v3"
)

func Routes(router fiber.Router) {
	router.Post("/chat_upload/authorize", CC.AuthorizeUpload)
	router.Post("/chat_upload/authorize/visual", CC.AuthorizeVisualUpload)

	router.Get("/chats", CC.GetChats)
	router.Delete("/chats/:partner_username", CC.DeleteChat)
}
