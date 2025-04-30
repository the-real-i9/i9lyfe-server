package chatRoute

import (
	CC "i9lyfe/src/controllers/chatControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Get("/chats", CC.GetChats)

	router.Delete("/chats/:partner_username", CC.DeleteChat)
}
