import express from "express"
import dotenv from "dotenv"

import PostCommentRoutes from "./private/postComment.routes.js"
import UserPrivateRoutes from "./private/user.private.routes.js"
import ChatRoutes from "./private/chat.routes.js"
import { expressSessionMiddleware } from "../middlewares/auth.middlewares.js"
import { verifyJwt } from "../services/security.services.js"

dotenv.config()

const router = express.Router()

router.use(
  expressSessionMiddleware(
    "session_store",
    process.env.SESSION_COOKIE_SECRET,
  ),
  (req, res, next) => {
    if (!req.session?.user) {
      return res.status(401).send("authentication required")
    }

    const { authJwt } = req.session.user

    try {
      req.auth = verifyJwt(authJwt, process.env.AUTH_JWT_SECRET)
    } catch (error) {
      return res.status(401).send(error)
    }

    next()
  }
)

router.use(PostCommentRoutes)
router.use(ChatRoutes)
router.use(UserPrivateRoutes)

export default router