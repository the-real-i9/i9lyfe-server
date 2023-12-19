import express from "express"
import expressSession from "express-session"
import pgSession from "connect-pg-simple"
import { Pool } from "pg"

import {
  emailVerificationController,
  newAccountRequestController,
  signinController,
  signupController,
} from "../controllers/authControllers.js"

const router = express.Router()

const PGStore = pgSession(expressSession)
router.use(
  "/signup",
  expressSession({
    store: new PGStore({
      pool: new Pool(),
      tableName: "ongoing_registration",
      createTableIfMissing: true,
    }),
    resave: false,
    saveUninitialized: false,
    // eslint-disable-next-line no-undef
    secret: process.env.SIGNUP_SESSION_COOKIE_SECRET,
    cookie: {
      maxAge: 1 * 60 * 60 * 1000,
      secure: false,
    }
  })
)

router.post("/signup/request_new_account", newAccountRequestController)

router.post("/signup/verify_email", emailVerificationController)

router.post("/signup/register_user", signupController)

router.post("/signin", signinController)

export default router
