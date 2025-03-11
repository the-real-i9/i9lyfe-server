import express from "express"

import * as signinControllers from "../controllers/auth/signin.controllers.js"
import * as passwordResetControllers from "../controllers/auth/passwordReset.controllers.js"
import * as signupControllers from "../controllers/auth/signup.controllers.js"
import * as signupValidators from "../validators/auth/signup.validators.js"
import * as signinValidators from "../validators/auth/signin.validators.js"
import * as passwordResetValidators from "../validators/auth/passwordReset.validators.js"
import { expressSession } from "../middlewares/auth.middlewares.js"

const router = express.Router()

router.use(expressSession())

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
