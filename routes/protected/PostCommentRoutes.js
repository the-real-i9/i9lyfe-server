import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as PCC from "../../controllers/PostCommentControllers.js"

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
router.post("/new_post", PCC.createNewPostController)
router.get("/home_feed", PCC.getHomeFeedController)
router.get("/posts/:post_id", PCC.getPostController)
router.delete("/posts/:post_id", PCC.deletePostController)

/* ====== POST'S REACTION ====== */

router.post("/react_to_post", PCC.createPostReactionController)
router.get("/posts/:post_id/reactors", PCC.getAllReactorsToPostController)
router.get(
  "/posts/:post_id/reactors/:reaction_code_point",
  PCC.getAllReactorsWithReactionToPostController
)
router.delete("/post_reactions/:post_id", PCC.removePostReactionController)

/* ====== POST'S COMMENT ====== */

router.post("/comment_on_post", PCC.createPostCommentController)
router.get("/posts/:post_id/comments", PCC.getAllCommentsOnPostController)
router.get("/comments/:comment_id", PCC.getCommentController)
router.delete("/post_comments/:comment_id", PCC.deletePostCommentController)

/* ====== COMMENT'S REACTION====== */

router.post("/react_to_comment", PCC.createCommentReactionController)
router.get("/comments/:comment_id/reactors", PCC.getAllReactorsToCommentController)
router.get(
  "/comments/:comment_id/reactors/:reaction_code_point",
  PCC.getAllReactorsWithReactionToCommentController
)
router.delete("/comment_reactions/:comment_id", PCC.removeCommentReactionController)

/* ====== COMMENT'S REPLY ====== */

router.post("/reply_to_comment", PCC.createCommentReplyController)
router.get("/comments/:comment_id/replies", PCC.getAllRepliesToCommentController)
router.get("/replies/:reply_id", PCC.getReplyController)
router.delete("/comment_replies/:reply_id", PCC.deleteCommentReplyController)

/* ====== REPOST ====== */

router.post("/repost", PCC.createRepostController)
router.delete("/reposts/:repost_id", PCC.deleteRepostController)

/* ====== POST SAVE ====== */

router.post("/save_post", PCC.postSaveController)
router.delete("/saved_posts/:post_id", PCC.postUnsaveController)

// GET insight data for a specific post

export default router
