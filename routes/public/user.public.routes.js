import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import {
  getUserFollowersController,
  getUserFollowingController,
  getUserPostsController,
  getUserProfileController,
} from "../../controllers/UserControllers.js"

const router = express.Router()

dotenv.config()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
    credentialsRequired: false,
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ error: err.inner.message })
    } else {
      next(err)
    }
  }
)

/* Users */
// GET a specific user's profile data
router.get("/:username", getUserProfileController)

// GET user followers
router.get("/:username/followers", getUserFollowersController)

// GET user followings
router.get("/:username/following", getUserFollowingController)

// GET user posts
router.get("/:username/posts", getUserPostsController)

export default router
