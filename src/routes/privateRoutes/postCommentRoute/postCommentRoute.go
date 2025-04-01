package postCommentRoute

import (
	PCC "i9lyfe/src/controllers/postCommentControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
/* ====== POST ====== */
router.Post("/new_post", PCC.CreateNewPost)
router.Get("/posts/:postId", PCC.GetPost)
router.Delete("/posts/:postId", PCC.DeletePost)

/* ====== POST'S REACTION ====== */

router.Post("/posts/:postId/react", PCC.ReactToPost)
router.Get(
  "/posts/:postId/reactors",
  ...validateParams,
  PCC.getReactorsToPost
)
router.Get(
  "/posts/:postId/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToPost
)
router.Delete(
  "/posts/:postId/remove_reaction",
  ...validateParams,
  PCC.removeReactionToPost
)

/* ====== POST'S COMMENT ====== */

router.Post(
  "/posts/:postId/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnPost
)

router.Get(
  "/posts/:postId/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnPost
)

router.Get("/comments/:commentId", ...validateParams, PCC.getComment)

router.Delete(
  "/posts/:postId/comments/:commentId",
  ...validateParams,
  PCC.removeCommentOnPost
)

router.Post(
  "/comments/:commentId/comment",
  ...validateParams,
  ...PCV.commentOn,
  PCC.commentOnComment
)
router.Get(
  "/comments/:commentId/comments",
  ...validateParams,
  ...validateLimitOffset,
  PCC.getCommentsOnComment
)

router.Delete(
  "/comments/:parentCommentId/comments/:commentId",
  ...validateParams,
  PCC.removeCommentOnComment
)

/* ====== COMMENT'S REACTION====== */

router.Post(
  "/comments/:commentId/react",
  ...validateParams,
  ...PCV.reactTo,
  PCC.reactToComment
)

router.Get(
  "/comments/:commentId/reactors",
  ...validateParams,
  PCC.getReactorsToComment
)

router.Get(
  "/comments/:commentId/reactors/:reaction",
  ...validateParams,
  PCC.getReactorsWithReactionToComment
)

router.Delete(
  "/comments/:commentId/remove_reaction",
  ...validateParams,
  PCC.removeReactionToComment
)

/* ====== REPOST ====== */

router.Post("/posts/:postId/repost", ...validateParams, PCC.createRepost)

/* ====== POST SAVE ====== */

router.Post("/posts/:postId/save", ...validateParams, PCC.savePost)
router.Delete("/posts/:postId/unsave", ...validateParams, PCC.unsavePost)
}
