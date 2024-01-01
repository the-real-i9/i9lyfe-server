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

/* All gets */

// GET all posts for a specific user ***. (Algorithmically aggregated)

// GET a post
router.get("/posts/:post_id")

// GET all comments on a post
router.get("/posts/:post_id/comments")

// GET a single comment on a post
router.get("/posts/:post_id/comments/:comment_id")

// GET all reactions to a post: returning all users that reacted to the post
router.get("/posts/:post_id/reactions")

// GET a single reaction to a post: limiting returned users to the ones with that reaction
router.get("/posts/:post_id/reactions/:reaction_code_point")

// GET all replies to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/posts/:post_id/comments/:comment_id/replies")

// GET a single reply to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
// the :reply_id is a single reply to the comment/reply with the that id
router.get("/posts/:post_id/comments/:comment_id/replies/:reply_id")

// GET all reactions to a comment/reply: returning all users that reacted to the comment
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/posts/:post_id/comments/:comment_id/reactions")

// GET a specific reaction to a comment/reply: limiting returned users to the ones with that reaction
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/posts/:post_id/comments/:comment_id/reactions/:reaction_code_point")

// GET insight data for a specific post

export default router
