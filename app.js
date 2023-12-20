import express from "express"
import expressSession from "express-session"
import pgSession from "connect-pg-simple"

import authRoutes from "./routes/authRoutes.js"
import { getDBPool } from "./models/db.js"

;(await import("dotenv")).config()

const app = express()

app.use(express.json())

const PGStore = pgSession(expressSession)
app.use(
  "/auth/signup",
  expressSession({
    store: new PGStore({
      pool: getDBPool(),
      tableName: "ongoing_registration",
      createTableIfMissing: true,
    }),
    resave: false,
    saveUninitialized: false,
    secret: process.env.SIGNUP_SESSION_COOKIE_SECRET,
    cookie: {
      maxAge: 1 * 60 * 60 * 1000,
      secure: false,
      path: "/auth/signup"
    },
  })
)

app.use("/auth", authRoutes)

export default app