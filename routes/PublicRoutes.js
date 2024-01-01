import express from "express"

const router = express.Router()

/* All gets: Public routes */

/* Users */
// GET a specific user's profile data
router.get("/:username")

// GET all posts for a specific user
router.get("/:username/posts")

/* Explore/Discover  */
// GET all explore/discover contents (all posts) aggregated algorithmically based on various stats. Basically, this is like the route Instagram makes call to to render its Explore/Discover

// GET all posts with search text

// GET photo posts with search text

// GET video posts with search text

// GET reel posts with search text

// GET story posts with search text

// GET hashtags with search text

/* Hashtags */
// GET all posts with an hashtag_name