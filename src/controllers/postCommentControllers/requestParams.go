package postCommentControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getPostParams struct {
	PostId string
}

func (p getPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getPostParams")
}

type deletePostParams struct {
	PostId string
}

func (p deletePostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "deletePostParams")
}

type reactToPostParams struct {
	PostId string
}

func (p reactToPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "reactToPostParams")
}

type getReactorsToPostParams struct {
	PostId string
}

func (p getReactorsToPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getReactorsToPostParams")
}

type getReactorsWithReactionToPostParams struct {
	PostId   string
	Reaction rune
}

func (p getReactorsWithReactionToPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
		validation.Field(&p.Reaction, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getReactorsWithReactionToPostParams")
}

type undoReactionToPostParams struct {
	PostId string
}

func (p undoReactionToPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "undoReactionToPostParams")
}

type commentOnPostParams struct {
	PostId string
}

func (p commentOnPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "commentOnPostParams")
}

type getCommentsOnPostParams struct {
	PostId string
}

func (p getCommentsOnPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getCommentsOnPostParams")
}

type getCommentParams struct {
	CommentId string
}

func (p getCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getCommentParams")
}

type removeCommentOnPostParams struct {
	PostId    string
	CommentId string
}

func (p removeCommentOnPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "removeCommentOnPostParams")
}

type reactToCommentParams struct {
	CommentId string
}

func (p reactToCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "reactToCommentParams")
}

type getReactorsToCommentParams struct {
	CommentId string
}

func (p getReactorsToCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getReactorsToCommentParams")
}

type getReactorsWithReactionToCommentParams struct {
	CommentId string
	Reaction  rune
}

func (p getReactorsWithReactionToCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
		validation.Field(&p.Reaction, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getReactorsWithReactionToCommentParams")
}

type undoReactionToCommentParams struct {
	CommentId string
}

func (p undoReactionToCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "undoReactionToCommentParams")
}

type commentOnCommentParams struct {
	CommentId string
}

func (p commentOnCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "commentOnCommentParams")
}

type getCommentsOnCommentParams struct {
	CommentId string
}

func (p getCommentsOnCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.CommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "getCommentsOnCommentParams")
}

type removeCommentOnCommentParams struct {
	ParentCommentId string
	ChildCommentId  string
}

func (p removeCommentOnCommentParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.ParentCommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
		validation.Field(&p.ChildCommentId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "removeCommentOnCommentParams")
}

type createRepostParams struct {
	PostId string
}

func (p createRepostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "createRepostParams")
}

type savePostParams struct {
	PostId string
}

func (p savePostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "savePostParams")
}

type undoSavePostParams struct {
	PostId string
}

func (p undoSavePostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestParams.go", "undoSavePostParams")
}
