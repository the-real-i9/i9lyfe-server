import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as PCC from "../../controllers/PostCommentControllers.js"
import { uploadCommentFiles, uploadPostFiles } from "../../middlewares/appMiddlewares.js"

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
router.get("/home_feed", PCC.getHomeFeedController)
router.get("/posts/:post_id", PCC.getPostController)
router.delete("/posts/:post_id", PCC.deletePostController)

/* ====== POST'S REACTION ====== */

router.post("/users/:user_id/posts/:post_id/react", PCC.reactToPostController)
router.get("/posts/:post_id/reactors", PCC.getReactorsToPostController)
router.get(
  "/posts/:post_id/reactors/:reaction",
  PCC.getReactorsWithReactionToPostController
)
router.delete("/posts/:post_id/remove_reaction", PCC.removeReactionToPostController)

/* ====== POST'S COMMENT ====== */

router.post("/users/:post_owner_user_id/posts/:post_id/comment", uploadCommentFiles, PCC.commentOnPostController)
router.post("/users/:parent_comment_owner_user_id/comments/:comment_id/comment", uploadCommentFiles, PCC.commentOnCommentController)

router.get("/posts/:post_id/comments", PCC.getCommentsOnPostController)
router.get("/comments/:parent_comment_id/comments", PCC.getCommentsOnCommentController)
router.get("/comments/:comment_id", PCC.getCommentController)

router.delete("/posts/:post_id/comments/:comment_id", PCC.deleteCommentOnPostController)
router.delete("/comments/:parent_comment_id/comments/:comment_id", PCC.deleteCommentOnCommentController)


/* ====== COMMENT'S REACTION====== */

router.post("/react_to_comment", PCC.reactToCommentController)
router.get("/comments/:comment_id/reactors", PCC.getReactorsToCommentController)
router.get(
  "/comments/:comment_id/reactors/:reaction_code_point",
  PCC.getReactorsWithReactionToCommentController
)
router.delete("/comments/:comment_id/remove_reaction", PCC.removeReactionToCommentController)

/* ====== REPOST ====== */

router.post("/repost", PCC.createRepostController)
router.delete("/reposts/:repost_id", PCC.deleteRepostController)

/* ====== POST SAVE ====== */

router.post("/save_post", PCC.postSaveController)
router.delete("/saved_posts/:post_id", PCC.postUnsaveController)

// GET insight data for a specific post

export default router
