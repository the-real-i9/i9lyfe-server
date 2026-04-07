package loginControllers

import (
	"i9lyfe/src/helpers"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type loginBody struct {
	EmailOrUsername string `msgpack:"emailOrUsername"`
	Password        string `msgpack:"password"`
}

func (b loginBody) Validate() error {

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

	return helpers.ValidationError(err, "loginControllers_requestBody.go", "loginBody")

}
