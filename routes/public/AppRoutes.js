import express from "express"
import {
  exploreController,
  getHashtagPostsController,
  searchFilterController,
} from "../../controllers/AppControllers.js"

const router = express.Router()

router.get("/explore", exploreController)

router.get("/explore/search", searchFilterController)

router.get("/hashtags/:hashtag_name", getHashtagPostsController)

export default router
