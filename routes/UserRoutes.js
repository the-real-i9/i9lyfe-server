import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import {
  followUserController,
  unfollowUserController,
  updateUserProfileController,
  uploadProfilePictureController,
} from "../controllers/UserControllers.js"

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

/* All gets: belonging to session user */
// GET user profile data
router.get("/me")

// GET user followers
router.get("/me/followers")
router.get("/:username/followers")

// GET user followings
router.get("/me/followings")
router.get("/:username/followings")

// GET user posts
router.get("/me/posts")

// GET posts user has been mentioned in
router.get("/me/mentions")
router.get("/:username/mentions")

// GET posts reacted by user
router.get("/me/likes")
router.get("/:user_id/mentions")

// GET posts saved by this user
router.get("/saved_posts")

// GET user notifications
router.get("/notifications")

export default router
