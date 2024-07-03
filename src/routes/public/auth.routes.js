import express from "express"
import dotenv from "dotenv"

import {
  expressSessionMiddleware,
  proceedEmailConfirmation,
  proceedEmailVerification,
  proceedPasswordReset,
  proceedUserRegistration,
} from "../../middlewares/auth.middlewares.js"
import { signinController } from "../../controllers/auth/signin.controller.js"
import { confirmEmailController, requestPasswordResetController, resetPasswordController } from "../../controllers/auth/passwordReset.controller.js"
import { registerUserController, requestNewAccountController, verifyEmailController } from "../../controllers/auth/signup.controller.js"

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

router.post("/signup/request_new_account", requestNewAccountController)
router.post("/signup/verify_email", proceedEmailVerification, verifyEmailController)
router.post("/signup/register_user", proceedUserRegistration, registerUserController)

router.post("/signin", signinController)

router.post("/forgot_password/request_password_reset", requestPasswordResetController)
router.post("/forgot_password/confirm_email", proceedEmailConfirmation, confirmEmailController)
router.post("/forgot_password/reset_password", proceedPasswordReset, resetPasswordController)

export default router
