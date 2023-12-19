import express from "express"

import {
  emailVerificationController,
  newAccountRequestController,
  signinController,
  signupController,
} from "../controllers/authControllers.js"

const router = express.Router()

router.post("/signup/request_new_account", newAccountRequestController)

router.post("/signup/verify_email", emailVerificationController)

router.post("/signup/register_user", signupController)

router.post("/signin", signinController)

export default router
