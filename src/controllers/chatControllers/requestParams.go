package chatControllers

import (
	"i9lyfe/src/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type deleteChatParams struct {
	PartnerUsername string
}

func (p deleteChatParams) Validate() error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PartnerUsername, validation.Required,
			validation.Length(2, 0).Error("username too short. minimum of 2 characters"),
			validation.Match(regexp.MustCompile("^\\w[\\w-]*\\w$")).Error("username contains invalid characters"),
		),
	)

	return helpers.ValidationError(err, "chatControllers_requestParams.go", "deleteChatParams")
}
