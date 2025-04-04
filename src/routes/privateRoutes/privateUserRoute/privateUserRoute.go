package privateUserRoute

import (
	UC "i9lyfe/src/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {

	router.Get("/client_user", UC.GetClientUser)

	router.Get("/signout", UC.Signout)

	router.Patch("/edit_profile", UC.EditUserProfile)

	router.Put("/change_profile_picture", UC.ChangeUserProfilePicture)

	router.Get("/home_feed", UC.GetHomeFeedPosts)

	router.Post("/users/:username/follow", UC.FollowUser)

	router.Delete("/users/:username/unfollow", UC.UnfollowUser)

	router.Get("/mentioned_posts", UC.GetUserMentionedPosts)

	router.Get("/reacted_posts", UC.GetUserReactedPosts)

	router.Get("/saved_posts", UC.GetUserSavedPosts)

	router.Get("/notifications", UC.GetUserNotifications)

	router.Put("/notifications/:notification_id/read", UC.ReadUserNotification)
}
