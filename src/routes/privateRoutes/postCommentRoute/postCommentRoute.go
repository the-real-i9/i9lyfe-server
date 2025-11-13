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
	router.Get("/posts/:postId/reactors", PCC.GetReactorsToPost)
	// router.Get("/posts/:postId/reactors/:reaction", PCC.GetReactorsWithReactionToPost)
	router.Delete("/posts/:postId/undo_reaction", PCC.RemoveReactionToPost)

	/* ====== POST'S COMMENT ====== */

	router.Post("/posts/:postId/comment", PCC.CommentOnPost)
	// remember to add this WARNINING in the APIdoc:
	// "for paginating comments, the 'offset' query should specify the date of the least recent item,
	// note that, the least recent item isn't always the first (ASC) or last (DESC) comment in the list retrieved.
	// the order of the comment list returned isn't evaluated based on date_created only
	// therefore, you shouldn't assume that the last or first item is the least recent item,
	// rather, you should programmatically find the least recent item in the list,
	// doing otherwise would cause items in previous pages to be returned and some items even missing
	router.Get("/posts/:postId/comments", PCC.GetCommentsOnPost)
	router.Get("/comments/:commentId", PCC.GetComment)
	router.Delete("/posts/:postId/comments/:commentId", PCC.RemoveCommentOnPost)

	/* ====== COMMENT'S REACTION====== */

	router.Post("/comments/:commentId/react", PCC.ReactToComment)
	router.Get("/comments/:commentId/reactors", PCC.GetReactorsToComment)
	// router.Get("/comments/:commentId/reactors/:reaction", PCC.GetReactorsWithReactionToComment)
	router.Delete("/comments/:commentId/undo_reaction", PCC.RemoveReactionToComment)

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
