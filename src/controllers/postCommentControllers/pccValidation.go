package postCommentControllers

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/cloudStorageService"
	"regexp"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type authorizePostUploadBody struct {
	PostType   string     `msgpack:"post_type" json:"post_type"`
	MediaMIME  [2]string  `msgpack:"media_mime" json:"media_mime"`
	MediaSizes [][2]int64 `msgpack:"media_sizes" json:"media_sizes"`
}

func (b authorizePostUploadBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PostType,
			validation.Required,
			validation.In("photo:portrait", "photo:square", "photo:landscape", "video:portrait", "video:square", "video:landscape", "reel").Error("invalid post type. expected 'photo:(portrait|square|landscape)', 'video:(portrait|square|landscape)', or 'reel'"),
		),

		validation.Field(&b.MediaMIME, validation.Required, validation.Length(2, 2).Error("expected array of 2 items."),
			validation.By(func(value any) error {
				val := value.([2]string)

				const (
					_             = iota
					BLUR_MIME int = iota - 1
					ACTUAL_MIME
				)

				if !slices.Contains([]string{"image/jpeg", "image/png", "image/webp", "image/avif"}, val[BLUR_MIME]) {
					return errors.New(`unsupported blur placeholder media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`)
				}

				if strings.HasPrefix(b.PostType, "photo") {
					if !slices.Contains([]string{"image/jpeg", "image/png", "image/webp", "image/avif"}, val[ACTUAL_MIME]) {
						return errors.New(`unsupported photo media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`)
					}
				} else {
					if !slices.Contains([]string{"video/mp4", "video/webm"}, val[ACTUAL_MIME]) {
						return errors.New(`unsupported video/reel media_mime; use one of ["video/mp4", "video/webm"]`)
					}
				}

				return nil
			}),
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

					if media_size[BLUR_PLACEHOLDER] < 1*1024 || media_size[BLUR_PLACEHOLDER] > 300*1024 {
						return errors.New("blur placeholder size out of range; min: 1KiB; max: 300KiB")
					}

					switch prefix, _, _ := strings.Cut(b.PostType, ":"); prefix {
					case "photo":
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 5*1024*1024 {
							return errors.New("photo media size out of range; min: 1KiB; max: 5MeB")
						}
					case "video":
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 15*1024*1024 {
							return errors.New("video or reel media size out of range; min: 1KiB; max: 15MeB")
						}
					default:
						if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 15*1024*1024 {
							return errors.New("reel media size out of range; min: 1KiB; max: 15MeB")
						}
					}

					return nil
				}),
			),
		),
	)

	return helpers.ValidationError(err, "pccValidation.go", "authorizePostUploadBody")
}

type createNewPostBody struct {
	MediaCloudNames []string `msgpack:"media_cloud_names"`
	Type            string   `msgpack:"type"`
	Description     string   `msgpack:"description"`
	At              int64    `msgpack:"at"`
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
				validation.Match(regexp.MustCompile(
					`^blur_placeholder:uploads/post/[\w-:/]+\w actual:uploads/post/[\w-:/]+\w$`,
				)).Error("invalid media cloud name"),
			),
		),
		validation.Field(&b.Description, validation.Length(0, 300)),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "pccValidation.go", "createNewPostBody")
	}

	for _, blurPlchActualMcn := range b.MediaCloudNames {
		go func(postType, blurPlchActualMcn string) {
			ctx := context.Background()

			var blurPlchMcn string
			var actualMcn string

			fmt.Sscanf(blurPlchActualMcn, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)

			if mInfo := cloudStorageService.GetMediaInfo(ctx, blurPlchMcn); mInfo != nil {
				if mInfo.Size < 1*1024 || mInfo.Size > 300*1024 {
					cloudStorageService.DeleteCloudMedia(ctx, blurPlchMcn)
				}
			}

			if mInfo := cloudStorageService.GetMediaInfo(ctx, actualMcn); mInfo != nil {
				switch prefix, _, _ := strings.Cut(postType, ":"); prefix {
				case "photo":
					if mInfo.Size < 1*1024 || mInfo.Size > 5*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, actualMcn)
					}
				case "video":
					if mInfo.Size < 1*1024 || mInfo.Size > 15*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, actualMcn)
					}
				default:
					if mInfo.Size < 1*1024 || mInfo.Size > 15*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, actualMcn)
					}
				}
			}

		}(b.Type, blurPlchActualMcn)
	}

	return nil
}

type reactToPostBody struct {
	Emoji string `msgpack:"emoji"`
	At    int64  `msgpack:"at"`
}

func (b reactToPostBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Emoji, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "pccValidation.go", "createNewPostBody")
}

type authorizeCommentUploadBody struct {
	AttachmentMIME string `msgpack:"attachment_mime" json:"attachment_mime"`
	AttachmentSize int64  `msgpack:"attachment_size" json:"attachment_size"`
}

func (b authorizeCommentUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.AttachmentMIME, validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif", "image/gif").Error(`unsupported attachment_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif", "image/gif"]`),
		),
		validation.Field(&b.AttachmentSize,
			validation.By(func(value any) error {
				val := value.(int64)

				if val < 1*1024 || val > 500*1024 {
					return errors.New("attachment_size out of range; min: 1KiB; max: 500KiB")
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "pccValidation.go", "authorizeCommentUploadBody")
}

type commentOnPostBody struct {
	CommentText         string `msgpack:"comment_text"`
	AttachmentCloudName string `msgpack:"attachment_cloud_name"`
	At                  int64  `msgpack:"at"`
}

func (b commentOnPostBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"),
				validation.Match(regexp.MustCompile(`^uploads/comment/[\w-/]+\w$`)).Error("invalid attachment cloud name"),
			)),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "pccValidation.go", "commentOnPostBody")
	}

	if b.AttachmentCloudName != "" {
		go func(attCn string) {
			ctx := context.Background()

			if mInfo := cloudStorageService.GetMediaInfo(ctx, attCn); mInfo != nil {
				if mInfo.Size < 1*1024 || mInfo.Size > 500*1024 {
					cloudStorageService.DeleteCloudMedia(ctx, attCn)
				}
			}

		}(b.AttachmentCloudName)
	}

	return nil
}

type reactToCommentBody struct {
	Emoji string `msgpack:"emoji"`
	At    int64  `msgpack:"at"`
}

func (b reactToCommentBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Emoji, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "pccValidation.go", "reactToCommentBody")
}

type commentOnCommentBody struct {
	CommentText         string `msgpack:"comment_text"`
	AttachmentCloudName string `msgpack:"attachment_cloud_name"`
	At                  int64  `msgpack:"at"`
}

func (b commentOnCommentBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.CommentText,
			validation.When(b.AttachmentCloudName == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"), validation.Length(0, 300).Error("comment_text length our of range. min:0. max:300"))),
		validation.Field(&b.AttachmentCloudName,
			validation.When(b.CommentText == "", validation.Required.Error("one of 'comment_text', 'attachment_cloud_name' or both must be provided"),
				validation.Match(regexp.MustCompile(`^uploads/comment/[\w-/]+\w$`)).Error("invalid attachment cloud name"),
			)),
		validation.Field(&b.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "pccValidation.go", "commentOnCommentBody")
	}

	if b.AttachmentCloudName != "" {
		go func(attCn string) {
			ctx := context.Background()

			if mInfo := cloudStorageService.GetMediaInfo(ctx, attCn); mInfo != nil {
				if mInfo.Size < 1*1024 || mInfo.Size > 500*1024 {
					cloudStorageService.DeleteCloudMedia(ctx, attCn)
				}
			}

		}(b.AttachmentCloudName)
	}

	return nil
}
