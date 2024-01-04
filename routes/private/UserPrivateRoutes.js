import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import {
  followUserController,
  unfollowUserController,
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

router.post("/follow_user", followUserController)

router.post("/unfollow_user", unfollowUserController)

router.put("/update_user_profile", updateUserProfileController)

router.put("/upload_profile_picture", uploadProfilePictureController)

// GET posts user has been mentioned in
router.get("/mentions")

// GET posts reacted to by user
router.get("/reacted_posts")

// GET posts saved by this user
router.get("/saved_posts")

// GET user notifications
router.get("/notifications")

export default router
