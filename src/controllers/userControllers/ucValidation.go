package userControllers

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/gcsHelpers"
	"regexp"

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

	return helpers.ValidationError(err, "ucValidation.go", "editProfileBody")
}

type authorizePPicUploadBody struct {
	PicMIME string   `json:"pic_mime"`
	PicSize [3]int64 `json:"pic_size"` // {small, medium, large}
}

func (b authorizePPicUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.PicMIME, validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported pic_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
		),
		validation.Field(&b.PicSize,
			validation.Required,
			validation.Length(3, 3).Error("expected an array of 3 items"),
			validation.By(func(value any) error {
				pic_size := value.([3]int64)

				const (
					_         = iota
					SMALL int = iota - 1
					MEDIUM
					LARGE
				)

				if pic_size[SMALL] < 1*1024 || pic_size[SMALL] > 500*1024 {
					return errors.New("small pic_size out of range; min: 1KiB; max: 500KiB")
				}

				if pic_size[MEDIUM] < 1*1024 || pic_size[MEDIUM] > 1*1024*1024 {
					return errors.New("medium pic_size out of range; min: 1KiB; max: 1MeB")
				}

				if pic_size[LARGE] < 1*1024 || pic_size[LARGE] > 2*1024*1024 {
					return errors.New("medium pic_size out of range; min: 1KiB; max: 2MeB")
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "ucValidation.go", "authorizePPicUploadBody")
}

type changeProfilePictureBody struct {
	ProfilePicCloudName string `json:"profile_pic_cloud_name"`
}

func (b changeProfilePictureBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.ProfilePicCloudName, validation.Required, validation.Match(regexp.MustCompile(
			`^small:uploads/user/profile_pics/[\w-/]+\w medium:uploads/user/profile_pics/[\w-/]+\w large:uploads/user/profile_pics/[\w-/]+\w$`,
		)).Error("invalid profile pic cloud name")),
	)

	if err != nil {
		return helpers.ValidationError(err, "ucValidation.go", "changeProfilePictureBody")
	}

	go func(ppicCn string) {
		ctx := context.Background()

		var (
			smallPPicCn  string
			mediumPPicCn string
			largePPicCn  string
		)

		fmt.Sscanf(ppicCn, "small:%s medium:%s large:%s", &smallPPicCn, &mediumPPicCn, &largePPicCn)

		if mInfo := gcsHelpers.GetMediaInfo(ctx, smallPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 500*1024 {
				gcsHelpers.DeleteCloudMedia(ctx, smallPPicCn)
			}
		}

		if mInfo := gcsHelpers.GetMediaInfo(ctx, mediumPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 1*1024*1024 {
				gcsHelpers.DeleteCloudMedia(ctx, mediumPPicCn)
			}
		}

		if mInfo := gcsHelpers.GetMediaInfo(ctx, largePPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 2*1024*1024 {
				gcsHelpers.DeleteCloudMedia(ctx, largePPicCn)
			}
		}
	}(b.ProfilePicCloudName)

	return nil
}
