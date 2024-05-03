import express from "express"
import dotenv from "dotenv"

import {
  expressSessionMiddleware,
  passwordResetProgressValidation,
  signupProgressValidation,
} from "../../middlewares/auth.middlewares.js"
import { signupController } from "../../controllers/auth/signupController.js"
import { signinController } from "../../controllers/auth/signin.controller.js"
import { passwordResetController } from "../../controllers/auth/passwordReset.controller.js"

dotenv.config()

const router = express.Router()

router.use(
  "/signup",
  expressSessionMiddleware(
    "ongoing_registration",
    process.env.SIGNUP_SESSION_COOKIE_SECRET,
    "/api/auth/signup"
  )
)

router.use(
  "/forgot_password",
  expressSessionMiddleware(
    "ongoing_password_reset",
    process.env.PASSWORD_RESET_SESSION_COOKIE_SECRET,
    "/api/auth/forgot_password"
  )
)

router.post("/signup/:step", signupProgressValidation, signupController)

router.post("/signin", signinController)

router.post("/forgot_password/:step", passwordResetProgressValidation, passwordResetController)

export default router
