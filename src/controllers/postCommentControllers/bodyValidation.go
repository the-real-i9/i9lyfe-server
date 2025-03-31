package postCommentControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type createNewPostBody struct {
	MediaDataList [][]byte `json:"media_data_list"`
	Type          string   `json:"type"`
	Description   string   `json:"description"`
}

func (b createNewPostBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.MediaDataList,
			validation.Required,
			validation.When(
				b.Type != "reel", validation.Length(1, 10).Error("media list out of range. minimum of 1 media item, maximum of 10 media items"),
			).Else(
				validation.Length(1, 1).Error("media list out of range for type 'reel'. must contain exactly 1 media item"),
			),
			validation.Each(
				validation.Length(100*1024, 8*(1024*1024)).Error("a media size is out of range. minimum of 100kb, maximum of 8 megabytes"),
			),
		),
		validation.Field(&b.Type,
			validation.Required,
			validation.In("photo", "video", "reel").Error("invalid post type. expected 'photo', 'video', or 'reel'"),
		),
		validation.Field(&b.Description, validation.Length(0, 300)),
	)

	return helpers.ValidationError(err, "postCommentControllers_bodyValidation.go", "createNewPostBody")
}
