import express from "express"
import dotenv from "dotenv"

import {
  expressSessionMiddleware,
  proceedEmailConfirmation,
  proceedEmailVerification,
  proceedPasswordReset,
  proceedUserRegistration,
} from "../../middlewares/auth.middlewares.js"
import signinController from "../../controllers/auth/signin.controller.js"
import * as passwordResetController from "../../controllers/auth/passwordReset.controller.js"
import * as signupController from "../../controllers/auth/signup.controller.js"
import * as authInputValidators from "../../middlewares/routeInputValidators/auth.inputValidators.js"

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

router.post("/signup/request_new_account", authInputValidators.requestNewAccount, signupController.requestNewAccount)
router.post("/signup/verify_email", authInputValidators.verifyEmail, proceedEmailVerification, signupController.verifyEmail)
router.post("/signup/register_user", authInputValidators.registerUser, proceedUserRegistration, signupController.registerUser)

router.post("/signin", authInputValidators.signin, signinController)

router.post("/forgot_password/request_password_reset", passwordResetController.requestPasswordReset)
router.post("/forgot_password/confirm_email", proceedEmailConfirmation, passwordResetController.confirmEmail)
router.post("/forgot_password/reset_password", proceedPasswordReset, passwordResetController.resetPassword)

export default router
