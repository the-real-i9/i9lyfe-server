import express from "express"

const router = express.Router()

/* Users */
// GET a specific user's profile data
router.get("/:username")

// GET user followers
router.get("/:username/followers")

// GET user followings
router.get("/:username/followings")

// GET user posts
router.get("/:username/posts")
