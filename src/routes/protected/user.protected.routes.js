import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import * as UC from "../../controllers/user.controllers.js"
import * as userValidators from "../../middlewares/validators/user.validators.js"

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


router.get("/home_feed", UC.getHomeFeed)

router.get("/session_user", UC.getSessionUser)

router.post("/users/:user_id/follow", ...userValidators.validateIdParams, UC.followUser)

router.delete("/users/:user_id/unfollow", ...userValidators.validateIdParams, UC.unfollowUser)

router.patch("/edit_profile", ...userValidators.editProfile, UC.editProfile)

router.put("/upload_profile_picture", UC.uploadProfilePicture)

router.patch("/update_connection_status", ...userValidators.updateConnectionStatus, UC.updateConnectionStatus)

router.put("/my_notifications/:notification_id/read", ...userValidators.validateIdParams, UC.readNotification)

// GET posts user has been mentioned in
router.get("/mentioned_posts", UC.getMentionedPosts)

// GET posts reacted to by user
router.get("/reacted_posts", UC.getReactedPosts)

// GET posts saved by this user
router.get("/saved_posts", UC.getSavedPosts)

// GET user notifications
router.get("/my_notifications", UC.getNotifications)

export default router
