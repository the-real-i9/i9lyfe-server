package privateUserRoute

import (
	UC "i9lyfe/src/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {

	router.Get("/client_user", UC.GetClientUser)

	router.Get("/me/signout", UC.Signout)

	router.Put("/me/edit_profile", UC.EditUserProfile)

	router.Put("/me/change_profile_picture", UC.ChangeUserProfilePicture)

	router.Post("/users/:username/follow", UC.FollowUser)

	router.Delete("/users/:username/unfollow", UC.UnfollowUser)

	router.Get("/me/mentioned_posts", UC.GetUserMentionedPosts)

	router.Get("/me/reacted_posts", UC.GetUserReactedPosts)

	router.Get("/me/saved_posts", UC.GetUserSavedPosts)

	router.Get("/me/notifications", UC.GetUserNotifications)

	router.Put("/me/notifications/:notification_id/read", UC.ReadUserNotification)
}
