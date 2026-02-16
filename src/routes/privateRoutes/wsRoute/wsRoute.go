// WebSocket Route
package wsRoute

import (
	"i9lyfe/src/controllers/realtimeController"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Use("/ws", func(c fiber.Ctx) error {
		if c.IsWebSocket() {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", realtimeController.WSStream)
}
