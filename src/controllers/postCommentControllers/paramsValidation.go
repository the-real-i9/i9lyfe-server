package postCommentControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getPostParams struct {
	PostId string `json:"post_id"`
}

func (b getPostParams) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PostId, validation.Required, is.UUID.Error("invalid id format. expected a UUID")),
	)

	return helpers.ValidationError(err, "postCommentControllers_paramsValidation.go", "getPostParams")
}
