package userControllers

import (
	"i9lyfe/src/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type followUserParams struct {
	Username string
}

func (p followUserParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "followUserParams")
}

type unfollowUserParams struct {
	Username string
}

func (p unfollowUserParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "unfollowUserParams")
}

type readUserNotificationParams struct {
	NotificationId string
}

func (p readUserNotificationParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.NotificationId, validation.Required, is.UUID.Error("invalid UUID string format")),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "readUserNotificationParams")
}

type getUserProfileParams struct {
	Username string
}

func (p getUserProfileParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "getUserProfileParams")
}

type getUserFollowersParams struct {
	Username string
}

func (p getUserFollowersParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "getUserFollowersParams")
}

type getUserFollowingParams struct {
	Username string
}

func (p getUserFollowingParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "getUserFollowingParams")
}

type getUserPostsParams struct {
	Username string
}

func (p getUserPostsParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Username, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_requestParams.go", "getUserPostsParams")
}
