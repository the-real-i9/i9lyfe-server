package publicUserRoutes

import (
	UC "i9lyfe/src/domain/user/userControllers"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Get("/:username", UC.GetUserProfile)

	router.Get("/:username/followers", UC.GetUserFollowers)

	router.Get("/:username/followings", UC.GetUserFollowings)

	router.Get("/:username/posts", UC.GetUserPosts)
}
