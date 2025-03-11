import express from "express"
import dotenv from "dotenv"

import UserPublicRoutes from "./public/user.public.routes.js"
import AppRoutes from "./public/app.routes.js"
import { verifyJwt } from "../services/security.services.js"
import { expressSession } from "../middlewares/auth.middlewares.js"

dotenv.config()

const router = express.Router()

router.use(
  expressSession(),
  (req, res, next) => {
    if (req.session?.user) {
      const { authJwt } = req.session.user
  
      try {
        req.auth = verifyJwt(authJwt)
      } catch (error) {
        return res.status(401).send(error)
      }
    }

    next()
  }
)


router.use(UserPublicRoutes)
router.use(AppRoutes)

export default router