import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import {
  createPostCommentController,
  createNewPostController,
  getAllReactorsToCommentController,
  getAllReactorsWithReactionToCommentController,
  getAllRepliesToCommentController,
  getAllCommentsOnPostController,
  getAllReactorsToPostController,
  getAllReactorsWithReactionToPostController,
  getReplyController,
  getCommentController,
  getPostController,
  createCommentReactionController,
  createPostReactionController,
  createCommentReplyController,
  createRepostController,
  createPostSaveController,
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

router.post("/new_post", createNewPostController)
router.delete("/posts/:post_id")

router.post("/post_reaction", createPostReactionController)
router.delete("/post_reactions/:post_id/:user_id")

router.post("/post_comment", createPostCommentController)
router.delete("/post_comments/:comment_id")

router.post("/comment_reaction", createCommentReactionController)
router.delete("/comment_reactions/:comment_id/:user_id")

router.post("/comment_reply", createCommentReplyController)
router.delete("/comment_replies/:reply_id")

router.post("/repost", createRepostController)
router.delete("/reposts/:repost_id")

router.post("/post_save", createPostSaveController)
router.delete("/post_saves/:post_id/:user_id")


/* All gets */

// GET all posts for a specific user ***. (Algorithmically aggregated)

// GET a post
router.get("/posts/:post_id", getPostController)

// GET all comments on a post
router.get("/posts/:post_id/comments", getAllCommentsOnPostController)

// GET single comment on a post
router.get(
  "/comments/:comment_id",
  getCommentController
)

// GET all reactions to a post: returning all users that reacted to the post
router.get("/posts/:post_id/reactors", getAllReactorsToPostController)

// GET a single reaction to a post: limiting returned users to the ones with that reaction
router.get(
  "/posts/:post_id/reactors/:reaction_code_point",
  getAllReactorsWithReactionToPostController
)

// GET all replies to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/comments/:comment_id/replies", getAllRepliesToCommentController)

// GET a single reply to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
// the :reply_id is a single reply to the comment/reply with the that id
router.get(
  "/replies/:reply_id",
  getReplyController
)

// GET all reactions to a comment/reply: returning all users that reacted to the comment
// the :comment_id either selects a comment or reply, since all replies are comments
router.get("/comments/:comment_id/reactors", getAllReactorsToCommentController)

// GET a specific reaction to a comment/reply: limiting returned users to the ones with that reaction
// the :comment_id either selects a comment or reply, since all replies are comments
router.get(
  "/comments/:comment_id/reactors/:reaction_code_point",
  getAllReactorsWithReactionToCommentController
)

// GET insight data for a specific post

export default router
