import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import {
  followUserController,
  getClientUserController,
  getUserMentionedPostsController,
  getUserNotificationsController,
  getUserReactedPostsController,
  getUserSavedPostsController,
  readUserNotificationController,
  unfollowUserController,
  updateUserConnectionStatusController,
  updateUserProfileController,
  uploadProfilePictureController,
} from "../../controllers/UserControllers.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ error: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.get("/client_user", getClientUserController)

router.post("/follow_user", followUserController)

router.delete("/followings/:followee_user_id", unfollowUserController)

router.put("/update_my_profile", updateUserProfileController)

router.put("/upload_profile_picture", uploadProfilePictureController)

router.put("/update_my_connection_status", updateUserConnectionStatusController)

router.put("/read_my_notification", readUserNotificationController)

// GET posts user has been mentioned in
router.get("/mentioned_posts", getUserMentionedPostsController)

// GET posts reacted to by user
router.get("/reacted_posts", getUserReactedPostsController)

// GET posts saved by this user
router.get("/saved_posts", getUserSavedPostsController)

// GET user notifications
router.get("/my_notifications", getUserNotificationsController)

export default router
