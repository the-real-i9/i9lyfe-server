import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as PCC from "../../controllers/postComment.controllers.js"
import { uploadCommentFiles, uploadPostFiles } from "../../middlewares/app.middlewares.js"

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
router.post("/new_post", uploadPostFiles, PCC.createNewPostController)
router.get("/posts/:post_id", PCC.getPostController)
router.delete("/posts/:post_id", PCC.deletePostController)

/* ====== POST'S REACTION ====== */

router.post("/users/:target_post_owner_user_id/posts/:target_post_id/react", PCC.reactToPostController)
router.get("/posts/:post_id/reactors", PCC.getReactorsToPostController)
router.get(
  "/posts/:post_id/reactors/:reaction",
  PCC.getReactorsWithReactionToPostController
)
router.delete("/posts/:target_post_id/remove_reaction", PCC.removeReactionToPostController)

/* ====== POST'S COMMENT ====== */

router.post("/users/:target_post_owner_user_id/posts/:target_post_id/comment", uploadCommentFiles, PCC.commentOnPostController)
router.post("/users/:target_comment_owner_user_id/comments/:target_comment_id/comment", uploadCommentFiles, PCC.commentOnCommentController)

router.get("/posts/:post_id/comments", PCC.getCommentsOnPostController)
router.get("/comments/:comment_id/comments", PCC.getCommentsOnCommentController)
router.get("/comments/:comment_id", PCC.getCommentController)

router.delete("/posts/:post_id/comments/:comment_id", PCC.removeCommentOnPostController)
router.delete("/comments/:parent_comment_id/comments/:comment_id", PCC.removeCommentOnCommentController)


/* ====== COMMENT'S REACTION====== */

router.post("/users/:target_comment_owner_user_id/comments/:target_comment_id/react", PCC.reactToCommentController)
router.get("/comments/:comment_id/reactors", PCC.getReactorsToCommentController)
router.get(
  "/comments/:comment_id/reactors/:reaction",
  PCC.getReactorsWithReactionToCommentController
)
router.delete("/comments/:target_comment_id/remove_reaction", PCC.removeReactionToCommentController)

/* ====== REPOST ====== */

router.post("/posts/:post_id/repost", PCC.createRepostController)
router.delete("/posts/:post_id/unrepost", PCC.deleteRepostController)

/* ====== POST SAVE ====== */

router.post("/posts/:post_id/save", PCC.postSaveController)
router.delete("/posts/:post_id/unsave", PCC.postUnsaveController)

// GET insight data for a specific post

export default router
