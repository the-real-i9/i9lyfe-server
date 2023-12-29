import express from "express"
import {
  commentOnPostController,
  createPostController,
  reactToCommentController,
  reactToPostController,
  replyToCommentController,
  repostPostController,
} from "../controllers/PostCommentControllers.js"

const router = express.Router()

router.post("/create_post", createPostController)

router.post("/react_to_post", reactToPostController)

router.post("/comment_on_post", commentOnPostController)

router.post("/react_to_comment", reactToCommentController)

router.post("/reply_to_comment", replyToCommentController)

router.post("/repost_post", repostPostController)

export default router
