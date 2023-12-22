import express from "express"
import dotenv from "dotenv"

import authRoutes from "./routes/authRoutes.js"
import { expressSessionMiddleware } from "./middlewares/authMiddlewares.js"

dotenv.config()

const app = express()

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
  "/auth/password_reset",
  expressSessionMiddleware(
    "ongoing_password_reset",
    process.env.PASSWORD_RESET_SESSION_COOKIE_SECRET,
    "/auth/password_reset"
  )
)

app.use("/auth", authRoutes)

export default app
