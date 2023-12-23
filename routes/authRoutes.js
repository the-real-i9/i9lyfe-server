import express from "express"

import { signupProgressValidation } from "../middlewares/authMiddlewares.js"
import { signupController } from "../controllers/authControllers/signupController.js"
import { signinController } from "../controllers/authControllers/signinController.js"
import { passwordResetController } from "../controllers/authControllers/passwordResetController.js"

const router = express.Router()

router.post("/signup", signupProgressValidation, signupController)

router.post("/signin", signinController)

router.post("/forgot_password", passwordResetController)

export default router
