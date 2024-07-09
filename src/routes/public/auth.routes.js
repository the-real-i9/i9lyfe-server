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
import * as PRC from "../../controllers/auth/passwordReset.controller.js"
import * as SC from "../../controllers/auth/signup.controller.js"
import * as AV from "../../middlewares/validators/auth.validators.js"

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

router.post("/signup/request_new_account", AV.requestNewAccount, SC.requestNewAccount)
router.post("/signup/verify_email", AV.verifyEmail, proceedEmailVerification, SC.verifyEmail)
router.post("/signup/register_user", AV.registerUser, proceedUserRegistration, SC.registerUser)

router.post("/signin", AV.signin, signinController)

router.post("/forgot_password/request_password_reset", AV.requestPasswordReset, PRC.requestPasswordReset)
router.post("/forgot_password/confirm_email", proceedEmailConfirmation, AV.confirmEmail, PRC.confirmEmail)
router.post("/forgot_password/reset_password", proceedPasswordReset, AV.resetPassword, PRC.resetPassword)

export default router
