import express from "express"

import * as UC from "../../controllers/user.controllers.js"
import { validateLimitOffset } from "../../validators/miscs.js"

const router = express.Router()

/* Users */
// GET a specific user's profile data
router.get("/:username", UC.getProfile)

// GET user followers
router.get("/:username/followers", ...validateLimitOffset, UC.getFollowers)

// GET user followings
router.get("/:username/following", ...validateLimitOffset, UC.getFollowings)

// GET user posts
router.get("/:username/posts", ...validateLimitOffset, UC.getPosts)

export default router
