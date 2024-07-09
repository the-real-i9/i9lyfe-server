import express from "express"
import dotenv from "dotenv"
import { expressjwt } from "express-jwt"
import * as AC from "../../controllers/app.controllers.js"
import * as AV from "../../middlewares/validators/app.validators.js"

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
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.get("/users/search", AC.searchUsersToChat)

router.get("/explore", AC.getExplorePosts)

router.get("/explore/search", AV.searchAndFilter, AC.searchAndFilter)

router.get("/hashtags/:hashtag_name", AC.getHashtagPosts)

export default router
