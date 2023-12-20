import express from "express"

import {
  emailVerificationController,
  newAccountRequestController,
  signinController,
  signupController,
} from "../controllers/authControllers.js"
import {
  confirmOngoingRegistration,
  rejectUnverifiedEmail,
  rejectVerifiedEmail,
} from "../middlewares/authMiddlewares.js"

const router = express.Router()

router.post("/signup/request_new_account", newAccountRequestController)

router.post(
  "/signup/verify_email",
  confirmOngoingRegistration,
  rejectVerifiedEmail,
  emailVerificationController
)

router.post(
  "/signup/register_user",
  confirmOngoingRegistration,
  rejectUnverifiedEmail,
  signupController
)

router.post("/signin", signinController)

export default router
