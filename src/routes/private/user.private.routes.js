import express from "express"

import * as UC from "../../controllers/user.controllers.js"
import * as userValidators from "../../validators/user.validators.js"
import {
  validateIdParams,
  validateLimitOffset,
} from "../../validators/miscs.js"

const router = express.Router()

router.get("/home_feed", ...validateLimitOffset, UC.getHomeFeedPosts)

router.get("/session_user", UC.getSessionUser)

router.post("/users/:user_id/follow", ...validateIdParams, UC.followUser)

router.delete("/users/:user_id/unfollow", ...validateIdParams, UC.unfollowUser)

router.patch("/edit_profile", ...userValidators.editProfile, UC.editProfile)

router.put(
  "/change_profile_picture",
  ...userValidators.changeProfilePicture,
  UC.changeProfilePicture
)

router.patch(
  "/update_connection_status",
  ...userValidators.updateConnectionStatus,
  UC.updateConnectionStatus
)

router.put(
  "/my_notifications/:notification_id/read",
  ...validateIdParams,
  UC.readNotification
)

// GET posts user has been mentioned in
router.get("/mentioned_posts", ...validateLimitOffset, UC.getMentionedPosts)

// GET posts reacted to by user
router.get("/reacted_posts", ...validateLimitOffset, UC.getReactedPosts)

// GET posts saved by this user
router.get("/saved_posts", ...validateLimitOffset, UC.getSavedPosts)

// GET user notifications
router.get(
  "/my_notifications",
  ...userValidators.getNotifications,
  UC.getNotifications
)

export default router
