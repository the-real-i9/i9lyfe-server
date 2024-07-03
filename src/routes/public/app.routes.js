import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"
import {
  getExplorePostsController,
  getHashtagPostsController,
  searchAndFilterController,
  searchUsersToChatController,
} from "../../controllers/app.controllers.js"

dotenv.config()

const router = express.Router()

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

router.get("/users/search", searchUsersToChatController)

router.get("/explore", getExplorePostsController)

router.get("/explore/search", searchAndFilterController)

router.get("/hashtags/:hashtag_name", getHashtagPostsController)

export default router
