import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/ChatControllers.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ error: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.get("/users_for_chat", CC.getUsersForChatController)

export default router