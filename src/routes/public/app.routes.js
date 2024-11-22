import express from "express"
import * as AC from "../../controllers/app.controllers.js"
import * as appValidators from "../../validators/app.validators.js"
import { validateLimitOffset } from "../../validators/miscs.js"


const router = express.Router()

router.get(
  "/users/search",
  ...appValidators.searchUsersToChat,
  AC.searchUsersToChat
)

router.get("/explore", ...validateLimitOffset, AC.getExplorePosts)

router.get(
  "/explore/search",
  ...appValidators.searchAndFilter,
  AC.searchAndFilter
)

router.get(
  "/hashtags/:hashtag_name",
  ...validateLimitOffset,
  AC.getHashtagPosts
)

export default router
