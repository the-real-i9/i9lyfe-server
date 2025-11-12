package publicUserRoute

import (
	UC "i9lyfe/src/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Get("/:username", UC.GetUserProfile)

	router.Get("/:username/followers", UC.GetUserFollowers)

	router.Get("/:username/followings", UC.GetUserFollowings)

	router.Get("/:username/posts", UC.GetUserPosts)
}
