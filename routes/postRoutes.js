import express from "express"
import { commentOnPostController, createPostController, reactToPostController } from "../controllers/postControllers.js"

const router = express.Router()

router.post("/create_post", createPostController)

router.post("/react_to_post", reactToPostController)

router.post("/comment_on_post", commentOnPostController)

export default router