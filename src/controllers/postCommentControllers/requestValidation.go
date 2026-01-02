package postCommentControllers

import (
	"i9lyfe/src/helpers"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type postMediaSize struct {
	BlurSize   int64 `json:"blur_size"`
	ActualSize int64 `json:"actual_size"`
}

func (b postMediaSize) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.ActualSize,
			validation.Required,
			validation.Min(1*1024*1024).Error("miminum of 1MeB for actual media size"),
			validation.Max(8*1024*1024).Error("maximum of 8MeB for actual media size"),
		),
		validation.Field(&b.BlurSize,
			validation.Required,
			validation.Min(1*1024).Error("miminum of 1KiB for blur media size"),
			validation.Max(10*1024).Error("maximum of 10KiB for blur media size"),
		),
	)
}

type authorizeUploadBody struct {
	PostType   string          `json:"post_type"`
	MediaMIME  string          `json:"media_mime"`
	MediaSizes []postMediaSize `json:"media_sizes"`
}

func (b authorizeUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.PostType,
			validation.Required,
			validation.In("photo:portrait", "photo:square", "photo:landscape", "video:portrait", "video:square", "video:landscape", "reel").Error("invalid post type. expected 'photo:(portrait|square|landscape)', 'video:(portrait|square|landscape)', or 'reel'"),
		),
		validation.Field(&b.MediaMIME, validation.Required,
			validation.When(strings.HasPrefix(b.PostType, "photo"), validation.Match(regexp.MustCompile(
				`^image/[a-zA-Z0-9][a-zA-Z0-9!#$&^_.+-]{0,126}(?:\s*;\s*[a-zA-Z0-9!#$&^_.+-]+=[^;]+)*$`,
			)).Error("expected mime_type to be a valid MIME type of the format image/*"),
			).Else(validation.Match(regexp.MustCompile(
				`^video/[a-zA-Z0-9][a-zA-Z0-9!#$&^_.+-]{0,126}(?:\s*;\s*[a-zA-Z0-9!#$&^_.+-]+=[^;]+)*$`,
			)).Error("expected mime_type to be a valid MIME type of the format video/*")),
		),
		validation.Field(&b.MediaSizes,
			validation.Required,
			validation.When(
				b.PostType != "reel", validation.Length(1, 10).Error("minimum of 1 media item, maximum of 10 media items"),
			).Else(
				validation.Length(1, 1).Error("must contain exactly 1 media item for type 'ree'"),
			),
			validation.Each(validation.Required),
		),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "authorizeUploadBody")

}

type createNewPostBody struct {
	MediaCloudNames []string `json:"media_cloud_names"`
	Type            string   `json:"type"`
	Description     string   `json:"description"`
	At              int64    `json:"at"`
}

func (b createNewPostBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Type,
			validation.Required,
			validation.In("photo:portrait", "photo:square", "photo:landscape", "video:portrait", "video:square", "video:landscape", "reel").Error("invalid post type. expected 'photo:(portrait|square|landscape)', 'video:(portrait|square|landscape)', or 'reel'"),
		),
		validation.Field(&b.MediaCloudNames,
			validation.Required,
			validation.When(
				b.Type != "reel", validation.Length(1, 10).Error("media list out of range. minimum of 1 media item, maximum of 10 media items"),
			).Else(
				validation.Length(1, 1).Error("media list out of range for type 'reel'. must contain exactly 1 media item"),
			),
		),
		validation.Field(&b.Description, validation.Length(0, 300)),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "createNewPostBody")
}

type reactToPostBody struct {
	Emoji string `json:"emoji"`
	At    int64  `json:"at"`
}

func (b reactToPostBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Emoji, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "createNewPostBody")
}

type commentOnPostBody struct {
	CommentText    string `json:"comment_text"`
	AttachmentData []byte `json:"attachment_data"`
	At             int64  `json:"at"`
}

func (b commentOnPostBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentData == nil, validation.Required.Error("one of 'comment_text', 'attachment_data' or both must be provided"), validation.Length(0, 300))),
		validation.Field(&b.AttachmentData,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_data' or both must be provided"), validation.Length(100*1024, 5*(1024*1024)).Error("attachment size is out of range. minimum of 100KiB, maximum of 5MiB"))),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "commentOnPostBody")
}

type reactToCommentBody struct {
	Emoji string `json:"emoji"`
	At    int64  `json:"at"`
}

func (b reactToCommentBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Emoji, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "reactToCommentBody")
}

type commentOnCommentBody struct {
	CommentText    string `json:"comment_text"`
	AttachmentData []byte `json:"attachment_data"`
	At             int64  `json:"at"`
}

func (b commentOnCommentBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentData == nil, validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(0, 300))),
		validation.Field(&b.AttachmentData,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_data' or both"), validation.Length(100*1024, 5*(1024*1024)).Error("attachment size is out of range. minimum of 100KiB, maximum of 5MiB"))),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "commentOnCommentBody")
}
