import express from "express"
import dotenv from "dotenv"

import {
  expressSessionMiddleware,
  passwordResetProgressValidation,
  signupProgressValidation,
} from "../../middlewares/authMiddlewares.js"
import { signupController } from "../../controllers/authControllers/signupController.js"
import { signinController } from "../../controllers/authControllers/signinController.js"
import { passwordResetController } from "../../controllers/authControllers/passwordResetController.js"

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

router.post("/signup", signupProgressValidation, signupController)

router.post("/signin", signinController)

router.post("/forgot_password", passwordResetProgressValidation, passwordResetController)

export default router
