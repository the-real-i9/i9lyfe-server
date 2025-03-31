package signupControllers

import (
	"i9lyfe/src/helpers"
	"regexp"
	"time"

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

	return helpers.ValidationError(err, "signupControllers_validation.go", "requestNewAccountBody")
}

type verifyEmailBody struct {
	Code string `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Code,
			validation.Required,
			validation.Length(6, 6).Error("expected a 6-digit number string"),
		),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "verifyEmailBody")
}

type registerUserBody struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
	Birthday time.Time `json:"birthday"`
	Bio      string    `json:"bio"`
}

func (b registerUserBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(3, 0).Error("username too short. minimum of 3 characters"),
			validation.Match(regexp.MustCompile("^[[:alnum:]][[:alnum:]_-]+[[:alnum:]]$")).Error("username contains invalid characters"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 0).Error("password too short. minimun of 8 characters"),
		),
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Birthday, validation.Required),
		validation.Field(&b.Bio, validation.Required, validation.Length(0, 150).Error("too many characters (max is 150)")),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "registerUserBody")
}
