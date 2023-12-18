import express from "express"
import {
  emailVerificationController,
  registrationRequestController,
  signinController,
  signupController,
} from "../controllers/auth_controllers.js"

const router = express.Router()

router.post("/registration_request", registrationRequestController)

router.post("/email_verification", emailVerificationController)

router.post("/signup", signupController)

router.post("/signin", signinController)

export default router
