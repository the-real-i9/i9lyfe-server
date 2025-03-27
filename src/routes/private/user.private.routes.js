import express from "express"

import * as UC from "../../controllers/user.controllers.js"
import * as userValidators from "../../validators/user.validators.js"
import {
  validateParams,
  validateLimitOffset,
} from "../../validators/miscs.js"

const router = express.Router()

router.get("/home_feed", ...validateLimitOffset, UC.getHomeFeedPosts)

router.get("/session_user", UC.getSessionUser)

router.get("/signout", UC.signout)

router.post("/users/:username/follow", ...validateParams, UC.followUser)

router.delete("/users/:username/unfollow", ...validateParams, UC.unfollowUser)

router.patch("/edit_profile", ...userValidators.editProfile, UC.editProfile)

router.put(
  "/change_profile_picture",
  ...userValidators.changeProfilePicture,
  UC.changeProfilePicture
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
  ...validateLimitOffset,
  UC.getNotifications
)

router.put(
  "/my_notifications/:notification_id/read",
  ...validateParams,
  UC.readNotification
)

export default router
