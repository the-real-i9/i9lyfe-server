package signupControllers

import (
	"i9lyfe/src/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestNewAccountBody struct {
	Email string `json:"email"`
}

func (b requestNewAccountBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.EmailFormat,
		),
	)

	return helpers.ValidationError(err, "signupControllers_requestBody.go", "requestNewAccountBody")
}

type verifyEmailBody struct {
	Code string `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Code, validation.Required),
	)

	return helpers.ValidationError(err, "signupControllers_requestBody.go", "verifyEmailBody")
}

type registerUserBody struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
	Name     string `json:"name"`
	Birthday int64  `json:"birthday"`
	Bio      string `json:"bio"`
}

func (b registerUserBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(2, 20).Error("username length out of range. min:2 max:20"),
			validation.Match(regexp.MustCompile(`^\w[\w-]*\w$`)).Error("username contains invalid characters"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 16).Error("password length out of range. min:8 max:16"),
		),
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Birthday, validation.Required),
		validation.Field(&b.Bio, validation.Required, validation.Length(0, 150).Error("bio too long (max chars: 150)")),
	)

	return helpers.ValidationError(err, "signupControllers_requestBody.go", "registerUserBody")
}
