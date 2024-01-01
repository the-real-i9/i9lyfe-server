import express from "express"
import {
  followUserController,
  updateUserProfileController,
  uploadProfilePictureController,
} from "../controllers/UserControllers.js"

const router = express.Router()

router.post("/follow_user", followUserController)

router.put("/update_user_profile", updateUserProfileController)

router.put("/upload_profile_picture", uploadProfilePictureController)

export default router
