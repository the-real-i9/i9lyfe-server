import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as PCC from "../../controllers/postComment.controllers.js"
import * as PCV from "../../middlewares/validators/postComment.validators.js"
import {
  uploadCommentFiles,
  uploadPostFiles,
} from "../../middlewares/app.middlewares.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

/* ====== POST ====== */
router.post("/new_post", ...PCV.createNewPost, uploadPostFiles, PCC.createNewPost)
router.get("/posts/:post_id", ...PCV.validateIdParams, PCC.getPost)
router.delete("/posts/:post_id", ...PCV.validateIdParams, PCC.deletePost)

/* ====== POST'S REACTION ====== */

router.post(
  "/users/:target_post_owner_user_id/posts/:target_post_id/react",
  ...PCV.validateIdParams,
  ...PCV.reactTo,
  PCC.reactToPost
)
router.get(
  "/posts/:post_id/reactors",
  ...PCV.validateIdParams,
  PCC.getReactorsToPost
)
router.get(
  "/posts/:post_id/reactors/:reaction",
  ...PCV.validateIdParams,
  PCC.getReactorsWithReactionToPost
)
router.delete(
  "/posts/:target_post_id/remove_reaction",
  ...PCV.validateIdParams,
  PCC.removeReactionToPost
)

/* ====== POST'S COMMENT ====== */

router.post(
  "/users/:target_post_owner_user_id/posts/:target_post_id/comment",
  ...PCV.validateIdParams,
  ...PCV.commentOn,
  uploadCommentFiles,
  PCC.commentOnPost
)
router.post(
  "/users/:target_comment_owner_user_id/comments/:target_comment_id/comment",
  ...PCV.validateIdParams,
  ...PCV.commentOn,
  uploadCommentFiles,
  PCC.commentOnComment
)

router.get(
  "/posts/:post_id/comments",
  ...PCV.validateIdParams,
  PCC.getCommentsOnPost
)
router.get(
  "/comments/:comment_id/comments",
  ...PCV.validateIdParams,
  PCC.getCommentsOnComment
)
router.get("/comments/:comment_id", ...PCV.validateIdParams, PCC.getComment)

router.delete(
  "/posts/:post_id/comments/:comment_id",
  ...PCV.validateIdParams,
  PCC.removeCommentOnPost
)
router.delete(
  "/comments/:parent_comment_id/comments/:comment_id",
  ...PCV.validateIdParams,
  PCC.removeCommentOnComment
)

/* ====== COMMENT'S REACTION====== */

router.post(
  "/users/:target_comment_owner_user_id/comments/:target_comment_id/react",
  ...PCV.validateIdParams,
  ...PCV.reactTo,
  PCC.reactToComment
)
router.get(
  "/comments/:comment_id/reactors",
  ...PCV.validateIdParams,
  PCC.getReactorsToComment
)
router.get(
  "/comments/:comment_id/reactors/:reaction",
  ...PCV.validateIdParams,
  PCC.getReactorsWithReactionToComment
)
router.delete(
  "/comments/:target_comment_id/remove_reaction",
  ...PCV.validateIdParams,
  PCC.removeReactionToComment
)

/* ====== REPOST ====== */

router.post("/posts/:post_id/repost", ...PCV.validateIdParams, PCC.createRepost)
router.delete(
  "/posts/:post_id/unrepost",
  ...PCV.validateIdParams,
  PCC.deleteRepost
)

/* ====== POST SAVE ====== */

router.post("/posts/:post_id/save", ...PCV.validateIdParams, PCC.postSave)
router.delete("/posts/:post_id/unsave", ...PCV.validateIdParams, PCC.postUnsave)

// GET insight data for a specific post

export default router
