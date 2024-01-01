import express from "express"
import dotenv from "dotenv"

import {
  followUserController,
  unfollowUserController,
  updateUserProfileController,
  uploadProfilePictureController,
} from "../controllers/UserControllers.js"
import { expressjwt } from "express-jwt"

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

export default router
