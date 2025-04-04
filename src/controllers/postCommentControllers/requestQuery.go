package postCommentControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getReactorsToPostQuery struct {
	Limit  int
	Offset int64
}

func (q getReactorsToPostQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getReactorsToPostQuery")
}

type getReactorsWithReactionToPostQuery struct {
	Limit  int
	Offset int64
}

func (q getReactorsWithReactionToPostQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getReactorsWithReactionToPostQuery")
}

type getCommentsOnPostQuery struct {
	Limit  int
	Offset int64
}

func (q getCommentsOnPostQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getCommentsOnPostQuery")
}

type getReactorsToCommentQuery struct {
	Limit  int
	Offset int64
}

func (q getReactorsToCommentQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getReactorsToCommentQuery")
}

type getReactorsWithReactionToCommentQuery struct {
	Limit  int
	Offset int64
}

func (q getReactorsWithReactionToCommentQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getReactorsWithReactionToCommentQuery")
}

type getCommentsOnCommentQuery struct {
	Limit  int
	Offset int64
}

func (q getCommentsOnCommentQuery) Validate() error {
	err := validation.ValidateStruct(&q,
		validation.Field(&q.Limit, validation.Min(1).Error("limit cannot be zero")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestQuery.go", "getCommentsOnCommentQuery")
}
