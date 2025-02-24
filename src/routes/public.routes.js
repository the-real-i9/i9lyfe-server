import express from "express"
import dotenv from "dotenv"

import AuthRoutes from "./public/auth.routes.js"
import UserPublicRoutes from "./public/user.public.routes.js"
import AppRoutes from "./public/app.routes.js"
import { verifyJwt } from "../services/security.services.js"
import { expressSessionMiddleware } from "../middlewares/auth.middlewares.js"

dotenv.config()

const router = express.Router()

router.use(
  "/app",
  expressSessionMiddleware(
    "user_session_private",
    process.env.USER_SESSION_COOKIE_SECRET,
    "/api/public/app",
    10 * 24 * 60 * 60 * 1000
  ),
  (req, res, next) => {
    if (req.session?.user) {
      const { authJwt } = req.session.user
  
      req.auth = verifyJwt(authJwt, process.env.JWT_SECRET)
    }

    return next()
  }
)

router.use("/auth", AuthRoutes)
router.use("/app", UserPublicRoutes)
router.use("/app", AppRoutes)

export default router