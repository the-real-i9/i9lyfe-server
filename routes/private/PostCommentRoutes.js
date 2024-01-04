import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import {
  commentOnPostController,
  createPostController,
  getAllCommentReactorsController,
  getAllCommentReactorsWithReactionController,
  getAllCommentRepliesController,
  getAllPostCommentsController,
  getAllPostReactorsController,
  getAllPostReactorsWithReactionController,
  getSingleCommentReplyController,
  getSinglePostCommentController,
  getSinglePostController,
  reactToCommentController,
  reactToPostController,
  replyToCommentController,
  repostPostController,
} from "../../controllers/PostCommentControllers.js"

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
router.get("/posts/:post_id", getSinglePostController)

// GET all comments on a post
router.get("/posts/:post_id/comments", getAllPostCommentsController)

// GET single comment on a post
router.get(
  "/comments/:comment_id",
  getSinglePostCommentController
)

// GET all reactions to a post: returning all users that reacted to the post
router.get("/posts/:post_id/reactors", getAllPostReactorsController)

// GET a single reaction to a post: limiting returned users to the ones with that reaction
router.get(
  "/posts/:post_id/reactors/:reaction_code_point",
  getAllPostReactorsWithReactionController
)

// GET all replies to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/comments/:comment_id/replies", getAllCommentRepliesController)

// GET a single reply to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
// the :reply_id is a single reply to the comment/reply with the that id
router.get(
  "/replies/:reply_id",
  getSingleCommentReplyController
)

// GET all reactions to a comment/reply: returning all users that reacted to the comment
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/comments/:comment_id/reactors", getAllCommentReactorsController)

// GET a specific reaction to a comment/reply: limiting returned users to the ones with that reaction
// the :comment_id either selects a comment or reply, since all replies are comments
router.get(
  "/comments/:comment_id/reactors/:reaction_code_point",
  getAllCommentReactorsWithReactionController
)

// GET insight data for a specific post

export default router
