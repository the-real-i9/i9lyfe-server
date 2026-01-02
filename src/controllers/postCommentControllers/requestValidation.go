package postCommentControllers

import (
	"errors"
	"i9lyfe/src/helpers"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type authorizeCommentUploadBody struct {
	AttachmentMIME string `json:"attachment_mime"`
	AttachmentSize int64  `json:"attachment_size"`
}

func (b authorizeCommentUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.AttachmentMIME, validation.Required,
			validation.Match(regexp.MustCompile(
				`^image/[a-zA-Z0-9][a-zA-Z0-9!#$&^_.+-]{0,126}(?:\s*;\s*[a-zA-Z0-9!#$&^_.+-]+=[^;]+)*$`,
			)).Error("expected attachment_mime to be a valid MIME type of format image/*"),
		),
		validation.Field(&b.AttachmentSize,
			validation.Min(1*1024).Error("attachment size too small. min: 1KiB"),
			validation.Max(10*1024).Error("attachment size too large. max: 10KiB"),
		),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "authorizeCommentUploadBody")
}

type authorizePostUploadBody struct {
	PostType string `json:"post_type"`
	// The first index gets the MIME for the blur frame of all media,
	// while the second index gets the MIME for the real media
	MediaMIME [2]string `json:"media_mime"`
	// The first index gets the size for the blur frame of a media,
	// while the second index gets the size for the real media
	MediaSizes [][2]int64 `json:"media_sizes"`
}

func (b authorizePostUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.PostType,
			validation.Required,
			validation.In("photo:portrait", "photo:square", "photo:landscape", "video:portrait", "video:square", "video:landscape", "reel").Error("invalid post type. expected 'photo:(portrait|square|landscape)', 'video:(portrait|square|landscape)', or 'reel'"),
		),
		validation.Field(&b.MediaMIME, validation.Required, validation.Length(2, 2).Error("expected array of 2 items.")),
		validation.Field(&b.MediaMIME[0], validation.Required,
			validation.Match(regexp.MustCompile(
				`^image/[a-zA-Z0-9][a-zA-Z0-9!#$&^_.+-]{0,126}(?:\s*;\s*[a-zA-Z0-9!#$&^_.+-]+=[^;]+)*$`,
			)).Error("expected media_mime for blur media to be a valid MIME type of the format image/*"),
		),
		validation.Field(&b.MediaMIME[1], validation.Required,
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
			validation.Each(
				validation.Length(2, 2).Error("expected arrays of 2 items each"),
				validation.By(func(value any) error {
					val := value.([2]int64)

					const (
						BLUR_FRAME int = 0
						REAL_MEDIA int = 1
					)

					if val[BLUR_FRAME] < 1*1024 || val[BLUR_FRAME] > 10*1024 {
						return errors.New("blur frame size out of range; min: 1KiB; max: 10KiB")
					}

					switch prefix, _, _ := strings.Cut(b.PostType, ":"); prefix {
					case "photo":
						if val[REAL_MEDIA] < 1*1024 || val[REAL_MEDIA] > 5*1024*1024 {
							return errors.New("real media size out of range; min: 1KiB; max: 5MeB")
						}
					case "video", "reel":
						if val[REAL_MEDIA] < 1*1024 || val[REAL_MEDIA] > 10*1024*1024 {
							return errors.New("real media size out of range; min: 1KiB; max: 10MeB")
						}
					default:
					}

					return nil
				}),
			),
		),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "authorizePostUploadBody")
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
	CommentText         string `json:"comment_text"`
	AttachmentCloudName string `json:"attachment_cloud_name"`
	At                  int64  `json:"at"`
}

func (b commentOnPostBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"))),
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
	CommentText         string `json:"comment_text"`
	AttachmentCloudName string `json:"attachment_cloud_name"`
	At                  int64  `json:"at"`
}

func (b commentOnCommentBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"))),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "commentOnCommentBody")
}
