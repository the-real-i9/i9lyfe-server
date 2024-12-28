import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import AuthRoutes from "./public/auth.routes.js"
import UserPublicRoutes from "./public/user.public.routes.js"
import AppRoutes from "./public/app.routes.js"

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

router.use("/auth", AuthRoutes)
router.use(UserPublicRoutes)
router.use(AppRoutes)

export default router