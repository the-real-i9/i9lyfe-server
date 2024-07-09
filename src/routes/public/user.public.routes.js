import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"

import * as UC from "../../controllers/user.controllers.js"

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
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

/* Users */
// GET a specific user's profile data
router.get("/:username", UC.getProfile)

// GET user followers
router.get("/:username/followers", UC.getFollowers)

// GET user followings
router.get("/:username/following", UC.getFollowing)

// GET user posts
router.get("/:username/posts", UC.getPosts)

export default router
