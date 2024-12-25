import express from "express"

import * as PCC from "../../controllers/postComment.controllers.js"
import * as PCV from "../../validators/postComment.validators.js"
import { validateIdParams } from "../../validators/miscs.js"

const router = express.Router()

/* ====== POST ====== */
router.post("/new_post", ...PCV.createNewPost, PCC.createNewPost)
router.get("/posts/:post_id", ...validateIdParams, PCC.getPost)
router.delete("/posts/:post_id", ...validateIdParams, PCC.deletePost)

/* ====== POST'S REACTION ====== */

router.post(
  "/posts/:post_id/react",
  ...validateIdParams,
  ...PCV.reactTo,
  PCC.reactToPost
)
router.get(
  "/posts/:post_id/reactors",
  ...validateIdParams,
  PCC.getReactorsToPost
)
router.get(
  "/posts/:post_id/reactors/:reaction",
  ...validateIdParams,
  PCC.getReactorsWithReactionToPost
)
router.delete(
  "/posts/:post_id/remove_reaction",
  ...validateIdParams,
  PCC.removeReactionToPost
)

/* ====== POST'S COMMENT ====== */

router.post(
  "/posts/:post_id/comment",
  ...validateIdParams,
  ...PCV.commentOn,
  PCC.commentOnPost
)
router.post(
  "/comments/:comment_id/comment",
  ...validateIdParams,
  ...PCV.commentOn,
  PCC.commentOnComment
)

router.get(
  "/posts/:post_id/comments",
  ...validateIdParams,
  PCC.getCommentsOnPost
)
router.get(
  "/comments/:comment_id/comments",
  ...validateIdParams,
  PCC.getCommentsOnComment
)
router.get("/comments/:comment_id", ...validateIdParams, PCC.getComment)

router.delete(
  "/posts/:post_id/comments/:comment_id",
  ...validateIdParams,
  PCC.removeCommentOnPost
)
router.delete(
  "/comments/:parent_comment_id/comments/:comment_id",
  ...validateIdParams,
  PCC.removeCommentOnComment
)

/* ====== COMMENT'S REACTION====== */

router.post(
  "/comments/:comment_id/react",
  ...validateIdParams,
  ...PCV.reactTo,
  PCC.reactToComment
)
router.get(
  "/comments/:comment_id/reactors",
  ...validateIdParams,
  PCC.getReactorsToComment
)
router.get(
  "/comments/:comment_id/reactors/:reaction",
  ...validateIdParams,
  PCC.getReactorsWithReactionToComment
)
router.delete(
  "/comments/:comment_id/remove_reaction",
  ...validateIdParams,
  PCC.removeReactionToComment
)

/* ====== REPOST ====== */

router.post("/posts/:post_id/repost", ...validateIdParams, PCC.createRepost)
router.delete("/posts/:post_id/unrepost", ...validateIdParams, PCC.deleteRepost)

/* ====== POST SAVE ====== */

router.post("/posts/:post_id/save", ...validateIdParams, PCC.savePost)
router.delete("/posts/:post_id/unsave", ...validateIdParams, PCC.unsavePost)

// GET insight data for a specific post

export default router
