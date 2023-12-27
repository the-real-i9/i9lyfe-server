import express from "express"
import { postCreationController, postReactionController } from "../controllers/postControllers.js"

const router = express.Router()

router.post("/create_post", postCreationController)

router.post("/react_to_post", postReactionController)

export default router