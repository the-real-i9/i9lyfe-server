package signinControllers

import (
	"i9lyfe/src/helpers"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type signInBody struct {
	EmailOrUsername string `json:"emailOrUsername"`
	Password        []byte `json:"password"`
}

func (b signInBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.EmailOrUsername,
			validation.Required,
			validation.When(strings.ContainsAny(b.EmailOrUsername, "@"),
				is.EmailFormat.Error("invalid email or username"),
			).Else(
				validation.Length(3, 0).Error("invalid email or username"),
			),
		),
		validation.Field(&b.Password,
			validation.Required,
		),
	)

	return helpers.ValidationError(err, "signinControllers_requestBody.go", "signInBody")

}
