import express from "express"
import * as AC from "../../controllers/app.controllers.js"
import * as appValidators from "../../validators/app.validators.js"
import { validateLimitOffset } from "../../validators/miscs.js"

const router = express.Router()

router.get("/explore_feed", ...validateLimitOffset, AC.getExploreFeed)

router.get("/explore_reels", ...validateLimitOffset, AC.getExploreReels)

router.get(
  "/explore/search",
  ...appValidators.searchAndFilter,
  AC.searchAndFilter
)

router.get(
  "/hashtags/:hashtag_name",
  ...appValidators.getHashtagPosts,
  AC.getHashtagPosts
)

export default router
