import express from "express"
import { postCreationController } from "../controllers/postControllers.js"

const router = express.Router()

router.post("/create_post", postCreationController)

export default router