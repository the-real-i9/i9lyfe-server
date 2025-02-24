import express from "express"
import dotenv from "dotenv"

import UserPublicRoutes from "./public/user.public.routes.js"
import AppRoutes from "./public/app.routes.js"
import { verifyJwt } from "../services/security.services.js"
import { expressSessionMiddleware } from "../middlewares/auth.middlewares.js"

dotenv.config()

const router = express.Router()

router.use(
  expressSessionMiddleware(
    "session_store",
    process.env.SESSION_COOKIE_SECRET,
  ),
  (req, res, next) => {
    if (req.session?.user) {
      const { authJwt } = req.session.user
  
      req.auth = verifyJwt(authJwt, process.env.AUTH_JWT_SECRET)
    }

    return next()
  }
)


router.use(UserPublicRoutes)
router.use(AppRoutes)

export default router