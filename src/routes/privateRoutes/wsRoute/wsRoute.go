// WebSocket Route
package wsRoute

import (
	"i9lyfe/src/controllers/realtimeController"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", realtimeController.WSStream)
}
