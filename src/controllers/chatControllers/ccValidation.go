package chatControllers

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

type authorizeUploadBody struct {
	MsgType   string `json:"msg_type"`
	MediaMIME string `json:"media_mime"`
	MediaSize int64  `json:"media_size"`
}

func (b authorizeUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.MsgType,
			validation.Required,
			validation.In("voice", "audio", "file").Error("invalid message type"),
		),
		validation.Field(&b.MediaMIME,
			validation.Required,
			validation.When(b.MsgType == "voice",
				validation.In("audio/ogg", "audio/aac").Error(`unsupported media_mime for voice; use one of ["audio/ogg", "audio/aac"]`),
			).Else(validation.When(b.MsgType == "audio",
				validation.In("audio/ogg", "audio/aac", "audio/mpeg", "audio/mp4", "audio/webm").Error(`unsupported media_mime for audio; use one of ["audio/ogg", "audio/aac", "audio/mpeg", "audio/mp4", "audio/webm"]`),
			).Else(validation.By(func(value any) error {
				val := value.(string)

				if strings.HasPrefix(val, "/") || strings.HasSuffix(val, "/") || !strings.Contains(val, "/") {
					return errors.New("invalid mime type")
				}

				return nil
			})),
			),
		),
		validation.Field(&b.MediaSize,
			validation.Required,
			validation.When(
				b.MsgType == "voice", validation.By(func(value any) error {
					val := value.(int64)

					if val < 500 || val > 10*1024*1024 {
						return errors.New("voice media_size out of range; min: 500BiB; max: 10MeB")
					}

					return nil
				}),
			).Else(validation.When(
				b.MsgType == "audio", validation.By(func(value any) error {
					val := value.(int64)

					if val < 500 || val > 20*1024*1024 {
						return errors.New("audio media_size out of range; min: 500BiB; max: 20MeB")
					}

					return nil
				}),
			).Else(validation.By(func(value any) error {
				val := value.(int64)

				if val < 500 || val > 50*1024*1024 {
					return errors.New("file media_size out of range; min: 500BiB; max: 50MeB")
				}

				return nil
			})),
			),
		),
	)

	return helpers.ValidationError(err, "ccValidation.go", "authorizeUploadBody")
}

type authorizeVisualUploadBody struct {
	MsgType   string    `json:"msg_type"`
	MediaMIME [2]string `json:"media_mime"`
	MediaSize [2]int64  `json:"media_size"`
}

func (b authorizeVisualUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.MsgType,
			validation.Required,
			validation.In("photo", "video").Error("invalid message type"),
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

				if b.MsgType == "photo" {
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
		validation.Field(&b.MediaSize,
			validation.Required,
			validation.Length(2, 2).Error("expected a media_size array of 2 items"),
			validation.By(func(value any) error {
				media_size := value.([2]int64)

				const (
					_                    = iota
					BLUR_PLACEHOLDER int = iota - 1
					ACTUAL_MEDIA
				)

				if media_size[BLUR_PLACEHOLDER] < 1*1024 || media_size[BLUR_PLACEHOLDER] > 100*1024 {
					return errors.New("blur placeholder media_size out of range; min: 1KiB; max: 100KiB")
				}

				switch b.MsgType {
				case "photo":
					if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 10*1024*1024 {
						return errors.New("photo media_size out of range; min: 1KiB; max: 10MeB")
					}
				case "video":
					if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 40*1024*1024 {
						return errors.New("video media_size out of range; min: 1KiB; max: 40MeB")
					}
				default:
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "ccValidation.go", "authorizeVisualUploadBody")
}

type MsgProps struct {
	TextContent    *string `json:"text_content,omitempty"`
	MediaCloudName *string `json:"media_cloud_name,omitempty"`
	Duration       *int64  `json:"duration,omitempty"`
	Caption        *string `json:"caption,omitempty"`
	Name           *string `json:"name,omitempty"`
}

type MsgContent struct {
	Type     string `json:"type"`
	MsgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "photo", "file").Error("invalid message type"),
		),
		validation.Field(&m.MediaCloudName,
			validation.When(m.Type == "text", validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.Match(regexp.MustCompile(
					`^blur_placeholder:uploads/chat/[\w-/]+\w actual:uploads/chat/[\w-/]+\w$`,
				)).Error("invalid media cloud name"),
			),
		),
		validation.Field(&m.MsgProps, validation.Required),
		validation.Field(&m.TextContent, validation.When(m.Type != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Duration, validation.When(m.Type != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file", "audio"}, m.Type), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(m.Type != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
	)

	if err != nil {
		return err
	}

	/* validate and (if needed) clean bad uploaded media */
	if mediaCloudName := m.MediaCloudName; mediaCloudName != nil {
		go func(msgType, mediaCloudName string) {

			ctx := context.Background()

			switch msgType {
			case "photo", "video":
				var (
					mcnBlur   string
					mcnActual string
				)

				fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &mcnBlur, &mcnActual)

				if mInfo := cloudStorageService.GetMediaInfo(ctx, mcnBlur); mInfo != nil {
					if mInfo.Size < 1*1024 || mInfo.Size > 100*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnBlur)
					}
				}

				if mInfo := cloudStorageService.GetMediaInfo(ctx, mcnActual); mInfo != nil {
					if msgType == "photo" && mInfo.Size < 1*1024 || mInfo.Size > 10*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnActual)
					} else if mInfo.Size < 1*1024 || mInfo.Size > 40*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnActual)
					}
				}
			case "voice":
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 10*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			case "audio":
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 20*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			default:
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 50*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			}
		}(m.Type, *mediaCloudName)
	}

	return nil
}

type sendMsgAcd struct {
	PartnerUsername  string     `json:"toUser"`
	IsReply          bool       `json:"isReply"`
	ReplyTargetMsgId string     `json:"replyTargetMsgId"`
	Msg              MsgContent `json:"msg"`
	At               int64      `json:"at"`
}

func (vb sendMsgAcd) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "sendMsgAcd")

}

type ackMsgDeliveredAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d ackMsgDeliveredAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "ackMsgDeliveredAcd")
}

type ackMsgReadAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d ackMsgReadAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "ackMsgReadAcd")
}

type getChatHistoryAcd struct {
	PartnerUsername string  `json:"partnerUsername"`
	Limit           int     `json:"limit"`
	Cursor          float64 `json:"cursor"`
}

func (d getChatHistoryAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "getChatHistoryAcd")
}

type reactToMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	Emoji           string `json:"emoji"`
	At              int64  `json:"at"`
}

func (d reactToMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.Emoji, validation.Required, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "reactToMsgAcd")
}

type removeReactionToMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d removeReactionToMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "removeReactionToMsgAcd")
}

type deleteMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	DeleteFor       string `json:"deleteFor"`
}

func (d deleteMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.DeleteFor, validation.Required, validation.In("me", "everyone").Error("expected value: 'me' or 'everyone'. but found "+d.DeleteFor)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "deleteMsgAcd")
}
