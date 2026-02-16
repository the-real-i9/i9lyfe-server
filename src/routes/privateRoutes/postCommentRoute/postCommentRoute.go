package postCommentRoute

import (
	PCC "i9lyfe/src/controllers/postCommentControllers"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Post("/post_upload/authorize", PCC.AuthorizePostUpload)
	router.Post("/comment_upload/authorize", PCC.AuthorizeCommentUpload)

	/* ====== POST ====== */
	router.Post("/new_post", PCC.CreateNewPost)
	router.Get("/posts/:postId", PCC.GetPost)
	router.Delete("/posts/:postId", PCC.DeletePost)

	/* ====== POST'S REACTION ====== */

	router.Post("/posts/:postId/react", PCC.ReactToPost)
	router.Get("/posts/:postId/reactors", PCC.GetReactorsToPost)
	// router.Get("/posts/:postId/reactors/:reaction", PCC.GetReactorsWithReactionToPost)
	router.Delete("/posts/:postId/remove_reaction", PCC.RemoveReactionToPost)

	/* ====== POST'S COMMENT ====== */

	router.Post("/posts/:postId/comment", PCC.CommentOnPost)
	router.Get("/posts/:postId/comments", PCC.GetCommentsOnPost)
	router.Get("/comments/:commentId", PCC.GetComment)
	router.Delete("/posts/:postId/comments/:commentId", PCC.RemoveCommentOnPost)

	/* ====== COMMENT'S REACTION====== */

	router.Post("/comments/:commentId/react", PCC.ReactToComment)
	router.Get("/comments/:commentId/reactors", PCC.GetReactorsToComment)
	// router.Get("/comments/:commentId/reactors/:reaction", PCC.GetReactorsWithReactionToComment)
	router.Delete("/comments/:commentId/remove_reaction", PCC.RemoveReactionToComment)

	/* ====== COMMENT'S COMMENT ===== */

	router.Post("/comments/:commentId/comment", PCC.CommentOnComment)
	router.Get("/comments/:commentId/comments", PCC.GetCommentsOnComment)
	router.Delete("/comments/:parentCommentId/comments/:childCommentId", PCC.RemoveCommentOnComment)

	/* ====== REPOST ====== */

	router.Post("/posts/:postId/repost", PCC.RepostPost)

	/* ====== POST SAVE ====== */

	router.Post("/posts/:postId/save", PCC.SavePost)
	router.Delete("/posts/:postId/unsave", PCC.UnsavePost)
}
