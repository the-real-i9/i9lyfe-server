import express from "express"

import * as PCC from "../../controllers/postComment.controllers.js"
import * as PCV from "../../validators/postComment.validators.js"
import { validateLimitOffset, validateParams } from "../../validators/miscs.js"

const router = express.Router()

/* ====== POST ====== */
router.post("/new_post", ...PCV.createNewPost, PCC.createNewPost)
router.get("/posts/:post_id", ...validateParams, PCC.getPost)
router.delete("/posts/:post_id", ...validateParams, PCC.deletePost)

/* ====== POST'S REACTION ====== */

router.post(
  "/posts/:post_id/react",
  ...validateParams,
  ...PCV.reactTo,
  PCC.reactToPost
)
router.get(
  "/posts/:post_id/reactors",
  ...validateParams,
  PCC.getReactorsToPost
)
router.get(
  "/posts/:post_id/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToPost
)
router.delete(
  "/posts/:post_id/remove_reaction",
  ...validateParams,
  PCC.removeReactionToPost
)

/* ====== POST'S COMMENT ====== */

router.post(
  "/posts/:post_id/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnPost
)

router.get(
  "/posts/:post_id/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnPost
)

router.get("/comments/:comment_id", ...validateParams, PCC.getComment)

router.delete(
  "/posts/:post_id/comments/:comment_id",
  ...validateParams,
  PCC.removeCommentOnPost
)

router.post(
  "/comments/:comment_id/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnComment
)
router.get(
  "/comments/:comment_id/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnComment
)

router.delete(
  "/comments/:parent_comment_id/comments/:comment_id",
  ...validateParams,
  PCC.removeCommentOnComment
)

/* ====== COMMENT'S REACTION====== */

router.post(
  "/comments/:comment_id/react",
  ...validateParams,
  ...PCV.reactTo,
  PCC.reactToComment
)

router.get(
  "/comments/:comment_id/reactors",
  ...validateParams,
  PCC.getReactorsToComment
)

router.get(
  "/comments/:comment_id/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToComment
)

router.delete(
  "/comments/:comment_id/remove_reaction",
  ...validateParams,
  PCC.removeReactionToComment
)

/* ====== REPOST ====== */

router.post("/posts/:post_id/repost", ...validateParams, PCC.createRepost)

/* ====== POST SAVE ====== */

router.post("/posts/:post_id/save", ...validateParams, PCC.savePost)
router.delete("/posts/:post_id/unsave", ...validateParams, PCC.unsavePost)

// GET insight data for a specific post

export default router
