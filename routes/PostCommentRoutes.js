import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import {
  commentOnPostController,
  createPostController,
  reactToCommentController,
  reactToPostController,
  replyToCommentController,
  repostPostController,
} from "../controllers/PostCommentControllers.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ error: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.post("/create_post", createPostController)

router.post("/react_to_post", reactToPostController)

router.post("/comment_on_post", commentOnPostController)

router.post("/react_to_comment", reactToCommentController)

router.post("/reply_to_comment", replyToCommentController)

router.post("/repost_post", repostPostController)

export default router
