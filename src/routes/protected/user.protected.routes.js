import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import * as UC from "../../controllers/user.controllers.js"
import * as userValidators from "../../middlewares/validators/user.validators.js"
import {
  validateIdParams,
  validateLimitOffset,
} from "../../middlewares/validators/miscs.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.get("/home_feed_posts", ...validateLimitOffset, UC.getHomeFeedPosts)

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
