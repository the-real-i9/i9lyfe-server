package passwordResetControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestPasswordResetBody struct {
	Email string `json:"email"`
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

type confirmActionBody struct {
	Token string `json:"token"`
}

func (b confirmActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Token, validation.Required),
	)

	return helpers.ValidationError(err, "passwordResetControllers_requestBody.go", "confirmActionBody")
}

type resetPasswordBody struct {
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
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
