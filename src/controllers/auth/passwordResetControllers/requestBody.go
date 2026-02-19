package passwordResetControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestPasswordResetBody struct {
	Email string `msgpack:"email" json:"email" format:"email"`
}

func (b requestPasswordResetBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.EmailFormat,
		),
	)

	return helpers.ValidationError(err, "passwordResetControllers_requestBody.go", "requestPasswordResetBody")
}

type confirmEmailBody struct {
	Token string `msgpack:"token" json:"token" minLength:"6" maxLength:"6"`
}

func (b confirmEmailBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Token, validation.Required),
	)

	return helpers.ValidationError(err, "passwordResetControllers_requestBody.go", "confirmEmailBody")
}

type resetPasswordBody struct {
	NewPassword        string `msgpack:"newPassword" json:"newPassword" format:"password"`
	ConfirmNewPassword string `msgpack:"confirmNewPassword" json:"confirmNewPassword" format:"password"`
}

func (b resetPasswordBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewPassword,
			validation.Required,
			validation.Length(8, 0).Error("password too short. minimun of 8 characters"),
		),
		validation.Field(&b.ConfirmNewPassword,
			validation.Required,
			validation.In(b.NewPassword).Error("passwords mismatch"),
		),
	)

	return helpers.ValidationError(err, "passwordResetControllers_requestBody.go", "resetPasswordBody")
}
