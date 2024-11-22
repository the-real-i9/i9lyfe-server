import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import PostCommentRoutes from "./private/postComment.routes.js"
import UserPrivateRoutes from "./private/user.private.routes.js"
import ChatRoutes from "./private/chat.routes.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.use(PostCommentRoutes)
router.use(ChatRoutes)
router.use(UserPrivateRoutes)

export default router