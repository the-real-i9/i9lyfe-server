import express from "express"
import dotenv from "dotenv"

import { expressSessionMiddleware } from "../../middlewares/auth.middlewares.js"
import * as signinControllers from "../../controllers/auth/signin.controllers.js"
import * as passwordResetControllers from "../../controllers/auth/passwordReset.controllers.js"
import * as signupControllers from "../../controllers/auth/signup.controllers.js"
import * as signupValidators from "../../validators/auth/signup.validators.js"
import * as signinValidators from "../../validators/auth/signin.validators.js"
import * as passwordResetValidators from "../../validators/auth/passwordReset.validators.js"

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

router.post(
  "/signup/request_new_account",
  ...signupValidators.requestNewAccount,
  signupControllers.requestNewAccount
)
router.post(
  "/signup/verify_email",
  ...signupValidators.verifyEmail,
  signupControllers.verifyEmail
)
router.post(
  "/signup/register_user",
  ...signupValidators.registerUser,
  signupControllers.registerUser
)

router.post("/signin", ...signinValidators.signin, signinControllers.signin)

router.post(
  "/forgot_password/request_password_reset",
  ...passwordResetValidators.requestPasswordReset,
  passwordResetControllers.requestPasswordReset
)
router.post(
  "/forgot_password/confirm_email",
  ...passwordResetValidators.confirmEmail,
  passwordResetControllers.confirmEmail
)
router.post(
  "/forgot_password/reset_password",
  ...passwordResetValidators.resetPassword,
  passwordResetControllers.resetPassword
)

export default router
