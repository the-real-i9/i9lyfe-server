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
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid id format. expected a UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_paramsValidation.go", "getPostParams")
}

type deletePostParams struct {
	PostId string
}

func (p deletePostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid id format. expected a UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_paramsValidation.go", "deletePostParams")
}

type reactToPostParams struct {
	PostId string
}

func (p reactToPostParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PostId, validation.Required, is.UUID.Error("invalid id format. expected a UUID string format")),
	)

	return helpers.ValidationError(err, "postCommentControllers_paramsValidation.go", "reactToPostParams")
}
