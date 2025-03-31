package postCommentRoutes

import (
	PCC "i9lyfe/src/controllers/postCommentControllers"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
/* ====== POST ====== */
router.Post("/new_post", PCC.CreateNewPost)
router.Get("/posts/:post_id", PCC.GetPost)
router.Delete("/posts/:post_id", ...validateParams, PCC.deletePost)

/* ====== POST'S REACTION ====== */

router.Post(
  "/posts/:post_id/react",
  ...validateParams,
  ...PCV.reactTo,
  PCC.reactToPost
)
router.Get(
  "/posts/:post_id/reactors",
  ...validateParams,
  PCC.getReactorsToPost
)
router.Get(
  "/posts/:post_id/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToPost
)
router.Delete(
  "/posts/:post_id/remove_reaction",
  ...validateParams,
  PCC.removeReactionToPost
)

/* ====== POST'S COMMENT ====== */

router.Post(
  "/posts/:post_id/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnPost
)

router.Get(
  "/posts/:post_id/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnPost
)

router.Get("/comments/:comment_id", ...validateParams, PCC.getComment)

router.Delete(
  "/posts/:post_id/comments/:comment_id",
  ...validateParams,
  PCC.removeCommentOnPost
)

router.Post(
  "/comments/:comment_id/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnComment
)
router.Get(
  "/comments/:comment_id/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnComment
)

router.Delete(
  "/comments/:parent_comment_id/comments/:comment_id",
  ...validateParams,
  PCC.removeCommentOnComment
)

/* ====== COMMENT'S REACTION====== */

router.Post(
  "/comments/:comment_id/react",
  ...validateParams,
  ...PCV.reactTo,
  PCC.reactToComment
)

router.Get(
  "/comments/:comment_id/reactors",
  ...validateParams,
  PCC.getReactorsToComment
)

router.Get(
  "/comments/:comment_id/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToComment
)

router.Delete(
  "/comments/:comment_id/remove_reaction",
  ...validateParams,
  PCC.removeReactionToComment
)

/* ====== REPOST ====== */

router.Post("/posts/:post_id/repost", ...validateParams, PCC.createRepost)

/* ====== POST SAVE ====== */

router.Post("/posts/:post_id/save", ...validateParams, PCC.savePost)
router.Delete("/posts/:post_id/unsave", ...validateParams, PCC.unsavePost)
}
