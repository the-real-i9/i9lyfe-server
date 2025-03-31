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

	return helpers.ValidationError(err, "passwordResetControllers_validation.go", "requestPasswordResetBody")
}

type confirmActionBody struct {
	Token string `json:"token"`
}

func (b confirmActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Token,
			validation.Required,
			validation.Length(6, 6).Error("expected a 6-digit number string"),
		),
	)

	return helpers.ValidationError(err, "passwordResetControllers_validation.go", "confirmActionBody")
}

type resetPasswordBody struct {
	NewPassword string `json:"newPassword"`
}

func (b resetPasswordBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewPassword,
			validation.Required,
			validation.Length(8, 0).Error("password too short. minimun of 8 characters"),
		),
	)

	return helpers.ValidationError(err, "passwordResetControllers_validation.go", "resetPasswordBody")
}
