import express from "express"
import { postCreationController } from "../controllers/post_managementControllers.js"

const router = express.Router()

router.post("/create_post", postCreationController)

export default router