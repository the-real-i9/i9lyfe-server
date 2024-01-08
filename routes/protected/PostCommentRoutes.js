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
  postSaveController,
  deletePostController,
  removePostReactionController,
  deletePostCommentController,
  removeCommentReactionController,
  deleteCommentReplyController,
  deleteRepostController,
  postUnsaveController,
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
/* ====== POST ====== */
router.post("/new_post", createNewPostController)
router.get("/posts/:post_id", getPostController)
router.delete("/posts/:post_id", deletePostController)

/* ====== POST'S REACTION ====== */

router.post("/post_reaction", createPostReactionController)
router.get("/posts/:post_id/reactors", getAllReactorsToPostController)
router.get(
  "/posts/:post_id/reactors/:reaction_code_point",
  getAllReactorsWithReactionToPostController
)
router.delete("/post_reactions/:post_id", removePostReactionController)

/* ====== POST'S COMMENT ====== */

router.post("/post_comment", createPostCommentController)
router.get("/posts/:post_id/comments", getAllCommentsOnPostController)
router.get("/comments/:comment_id", getCommentController)
router.delete("/post_comments/:comment_id", deletePostCommentController)

/* ====== COMMENT'S REACTION====== */

router.post("/comment_reaction", createCommentReactionController)
router.get("/comments/:comment_id/reactors", getAllReactorsToCommentController)
router.get(
  "/comments/:comment_id/reactors/:reaction_code_point",
  getAllReactorsWithReactionToCommentController
)
router.delete("/comment_reactions/:comment_id", removeCommentReactionController)

/* ====== COMMENT'S REPLY ====== */

router.post("/comment_reply", createCommentReplyController)
router.get("/comments/:comment_id/replies", getAllRepliesToCommentController)
router.get("/replies/:reply_id", getReplyController)
router.delete("/comment_replies/:reply_id", deleteCommentReplyController)

/* ====== REPOST ====== */

router.post("/repost", createRepostController)
router.delete("/reposts/:repost_id", deleteRepostController)

/* ====== POST SAVE ====== */

router.post("/post_save", postSaveController)
router.delete("/post_saves/:post_id", postUnsaveController)

// GET insight data for a specific post

export default router
