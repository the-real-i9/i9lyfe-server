package userControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type editProfileBody struct {
	Name     string `json:"name"`
	Birthday int64  `json:"birthday"`
	Bio      string `json:"bio"`
}

func (b editProfileBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required.When(b.Birthday == 0 && b.Bio == "").Error("no field provided. at least one field must be provided")),
		validation.Field(&b.Birthday, validation.Required.When(b.Name == "" && b.Bio == "").Error("no field provided. at least one field must be provided")),
		validation.Field(&b.Bio, validation.Required.When(b.Birthday == 0 && b.Name == "").Error("no field provided. at least one field must be provided"), validation.Length(0, 150).Error("too many characters (max is 150)")),
	)

	return helpers.ValidationError(err, "userControllers_requestBody.go", "editProfileBody")
}

type changeProfilePictureBody struct {
	PictureData []byte `json:"picture_data"`
}

func (b changeProfilePictureBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PictureData, validation.Required, validation.Length(1024, 2*(1024*1024)).Error("picture size is out of range. minimum of 1KiB, maximum of 2MiB")),
	)

	return helpers.ValidationError(err, "userControllers_requestBody.go", "changeProfilePictureBody")
}
