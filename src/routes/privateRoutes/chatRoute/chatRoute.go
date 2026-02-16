package chatRoute

import (
	CC "i9lyfe/src/controllers/chatControllers"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Post("/chat_upload/authorize", CC.AuthorizeUpload)
	router.Post("/chat_upload/authorize/visual", CC.AuthorizeVisualUpload)

	router.Get("/chats", CC.GetChats)
	router.Delete("/chats/:partner_username", CC.DeleteChat)
}
