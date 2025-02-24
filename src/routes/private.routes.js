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
  "/app",
  expressSessionMiddleware(
    "user_session_private",
    process.env.USER_SESSION_COOKIE_SECRET,
    "/api/private/app",
    10 * 24 * 60 * 60 * 1000
  ),
  (req, res, next) => {
    if (!req.session?.user) {
      return res.status(401).send("authentication required")
    }

    const { authJwt } = req.session.user

    req.auth = verifyJwt(authJwt, process.env.JWT_SECRET)

    return next()
  }
)

router.use("/app", PostCommentRoutes)
router.use("/app", ChatRoutes)
router.use("/app", UserPrivateRoutes)

export default router