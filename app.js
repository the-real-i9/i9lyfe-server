import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import authRoutes from "./routes/authRoutes.js"
import PostCommentRoutes from "./routes/PostCommentRoutes.js"
import { expressSessionMiddleware } from "./middlewares/authMiddlewares.js"

dotenv.config()

const app = express()

app.use(cors())

app.use(express.json())

app.use(
  "/auth/signup",
  expressSessionMiddleware(
    "ongoing_registration",
    process.env.SIGNUP_SESSION_COOKIE_SECRET,
    "/auth/signup"
  )
)

app.use(
  "/auth/forgot_password",
  expressSessionMiddleware(
    "ongoing_password_reset",
    process.env.PASSWORD_RESET_SESSION_COOKIE_SECRET,
    "/auth/forgot_password"
  )
)

app.use("/auth", authRoutes)
app.use(PostCommentRoutes)

export default app
