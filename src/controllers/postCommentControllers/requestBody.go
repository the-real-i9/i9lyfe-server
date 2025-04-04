package postCommentControllers

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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
				validation.Length(100*1024, 8*(1024*1024)).Error("one or more media size is out of range. minimum of 100KiB, maximum of 8MiB"),
			),
		),
		validation.Field(&b.Type,
			validation.Required,
			validation.In("photo", "video", "reel").Error("invalid post type. expected 'photo', 'video', or 'reel'"),
		),
		validation.Field(&b.Description, validation.Length(0, 300)),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestBody.go", "createNewPostBody")
}

type reactToPostBody struct {
	Reaction rune `json:"reaction"`
}

func (b reactToPostBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Reaction, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestBody.go", "createNewPostBody")
}

type commentOnPostBody struct {
	CommentText    string `json:"comment_text"`
	AttachmentData []byte `json:"attachment_data"`
}

func (b commentOnPostBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentData == nil, validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(0, 300))),
		validation.Field(&b.AttachmentData,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(100*1024, 5*(1024*1024)).Error("attachment size is out of range. minimum of 100KiB, maximum of 5MiB"))),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestBody.go", "commentOnPostBody")
}

type reactToCommentBody struct {
	Reaction rune `json:"reaction"`
}

func (b reactToCommentBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Reaction, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestBody.go", "reactToCommentBody")
}

type commentOnCommentBody struct {
	CommentText    string `json:"comment_text"`
	AttachmentData []byte `json:"attachment_data"`
}

func (b commentOnCommentBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentData == nil, validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(0, 300))),
		validation.Field(&b.AttachmentData,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(100*1024, 5*(1024*1024)).Error("attachment size is out of range. minimum of 100KiB, maximum of 5MiB"))),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestBody.go", "commentOnCommentBody")
}
