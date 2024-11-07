import express from "express"
import dotenv from "dotenv"

import {
  expressSessionMiddleware,
  proceedEmailConfirmation,
  proceedEmailVerification,
  proceedPasswordReset,
  proceedUserRegistration,
} from "../../middlewares/auth.middlewares.js"
import signinController from "../../controllers/auth/signin.auth.controllers.js"
import * as PRC from "../../controllers/auth/passwordReset.auth.controllers.js"
import * as SC from "../../controllers/auth/signup.auth.controllers.js"
import * as authValidators from "../../middlewares/validators/auth.validators.js"

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

router.post("/signup/request_new_account", ...authValidators.requestNewAccount, SC.requestNewAccount)
router.post("/signup/verify_email", ...authValidators.verifyEmail, proceedEmailVerification, SC.verifyEmail)
router.post("/signup/register_user", ...authValidators.registerUser, proceedUserRegistration, SC.registerUser)

router.post("/signin", authValidators.signin, signinController)

router.post("/forgot_password/request_password_reset", authValidators.requestPasswordReset, PRC.requestPasswordReset)
router.post("/forgot_password/confirm_email", proceedEmailConfirmation, authValidators.confirmEmail, PRC.confirmEmail)
router.post("/forgot_password/reset_password", proceedPasswordReset, authValidators.resetPassword, PRC.resetPassword)

export default router
