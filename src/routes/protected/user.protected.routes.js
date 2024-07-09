import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import * as userControllers from "../../controllers/user.controllers.js"
import * as userValidators from "../../middlewares/inputValidators/user.validators.js"

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


router.get("/home_feed", userControllers.getHomeFeed)

router.get("/session_user", userControllers.getSessionUser)

router.post("/users/:user_id/follow", userControllers.followUser)

router.delete("/users/:user_id/unfollow", userControllers.unfollowUser)

router.patch("/edit_profile", userValidators.editProfile, userControllers.editProfile)

router.put("/upload_profile_picture", userControllers.uploadProfilePicture)

router.patch("/update_connection_status", userValidators.updateConnectionStatus, userControllers.updateConnectionStatus)

router.put("/my_notifications/:notification_id/read", userControllers.readNotification)

// GET posts user has been mentioned in
router.get("/mentioned_posts", userControllers.getMentionedPosts)

// GET posts reacted to by user
router.get("/reacted_posts", userControllers.getReactedPosts)

// GET posts saved by this user
router.get("/saved_posts", userControllers.getSavedPosts)

// GET user notifications
router.get("/my_notifications", userControllers.getNotifications)

export default router
