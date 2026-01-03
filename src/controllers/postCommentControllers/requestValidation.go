package postCommentControllers

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
)

type authorizeCommentUploadBody struct {
	AttachmentMIME string `json:"attachment_mime"`
	AttachmentSize int64  `json:"attachment_size"`
}

func (b authorizeCommentUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.AttachmentMIME, validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported attachment_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
		),
		validation.Field(&b.AttachmentSize,
			validation.By(func(value any) error {
				val := value.(int64)

				if val < 1024 || val > 10*1024 {
					return errors.New("attachment_size out of range; min: 1KiB; max: 10KiB")
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "authorizeCommentUploadBody")
}

type authorizePostUploadBody struct {
	PostType   string     `json:"post_type"`
	MediaMIME  [2]string  `json:"media_mime"`  // {blur_placeholder, actual}
	MediaSizes [][2]int64 `json:"media_sizes"` // {{blur_placeholder, actual}, ...}
}

func (b authorizePostUploadBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PostType,
			validation.Required,
			validation.In("photo:portrait", "photo:square", "photo:landscape", "video:portrait", "video:square", "video:landscape", "reel").Error("invalid post type. expected 'photo:(portrait|square|landscape)', 'video:(portrait|square|landscape)', or 'reel'"),
		),
		validation.Field(&b.MediaMIME, validation.Required, validation.Length(2, 2).Error("expected array of 2 items.")),
		validation.Field(&b.MediaMIME[0], validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported blur placeholder media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
		),
		validation.Field(&b.MediaMIME[1], validation.Required,
			validation.When(strings.HasPrefix(b.PostType, "photo"),
				validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported photo media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
			).Else(
				validation.In("video/mp4", "video/webm").Error(`unsupported video/reel media_mime; use one of ["video/mp4", "video/webm"]`),
			),
		),
		validation.Field(&b.MediaSizes,
			validation.Required,
			validation.When(
				b.PostType != "reel", validation.Length(1, 10).Error("minimum of 1 media item, maximum of 10 media items"),
			).Else(
				validation.Length(1, 1).Error("must contain exactly 1 media item for type 'ree'"),
			),
			validation.Each(
				validation.Length(2, 2).Error("expected a media_size array of 2 items"),
				validation.By(func(value any) error {
					media_size := value.([2]int64)

					const (
						_                    = iota
						BLUR_PLACEHOLDER int = iota - 1
						ACTUAL_MEDIA
					)

					if media_size[BLUR_PLACEHOLDER] < 1*1024 || media_size[BLUR_PLACEHOLDER] > 10*1024 {
						return errors.New("blur placeholder size out of range; min: 1KiB; max: 10KiB")
					}

					switch prefix, _, _ := strings.Cut(b.PostType, ":"); prefix {
					case "photo":
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 5*1024*1024 {
							return errors.New("photo media size out of range; min: 1KiB; max: 5MeB")
						}
					case "video":
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 15*1024*1024 {
							return errors.New("video or reel media size out of range; min: 1KiB; max: 10MeB")
						}
					case "reel":
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 10*1024*1024 {
							return errors.New("reel media size out of range; min: 1KiB; max: 10MeB")
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

func (b createNewPostBody) Validate(ctx context.Context) error {

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
			validation.Each(
				validation.By(func(value any) error {
					val := value.(string)

					if !(strings.HasPrefix(val, "blur_placeholder:uploads/post/") && strings.Contains(val, " actual:uploads/post/")) {
						return errors.New("invalid media cloud name")
					}

					return nil
				}),
			),
		),
		validation.Field(&b.Description, validation.Length(0, 300)),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "createNewPostBody")
	}

	for _, blurPlchActualMcn := range b.MediaCloudNames {
		var blurPlchMcn string
		var actualMcn string

		fmt.Sscanf(blurPlchActualMcn, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)

		for _, mcn := range []string{blurPlchMcn, actualMcn} {
			_, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Attrs(ctx)
			if errors.Is(err, storage.ErrObjectNotExist) {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: media (%s) does not exist in cloud", mcn))
			}
		}
	}

	return nil
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

func (b commentOnPostBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"))),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "commentOnPostBody")
	}

	if b.AttachmentCloudName != "" {
		_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(b.AttachmentCloudName).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: attachment (%s) does not exist in cloud", b.AttachmentCloudName))
		}
	}

	return nil
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

func (b commentOnCommentBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"))),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "postCommentControllers_requestValidation.go", "commentOnCommentBody")
	}

	if b.AttachmentCloudName != "" {
		_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(b.AttachmentCloudName).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: attachment (%s) does not exist in cloud", b.AttachmentCloudName))
		}
	}

	return nil
}
