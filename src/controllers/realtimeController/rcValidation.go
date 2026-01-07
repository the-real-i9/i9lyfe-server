package realtimeController

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/gcsHelpers"
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rtActionBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b rtActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action, validation.Required),
		validation.Field(&b.Data,
			validation.Required.When(
				!slices.Contains([]string{
					"subscribe to live content metrics",
					"unsubscribe from live content metrics",
					"subscribe to user presence change",
					"unsubscribe from user presence change",
				}, b.Action),
			),
		),
	)

	return helpers.ValidationError(err, "rcValidation.go", "rtActionBody")
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

type subToUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb subToUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "subToUserPresenceAcd")
}

type unsubFromUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb unsubFromUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "unsubFromUserPresenceAcd")
}

type MsgProps struct {
	TextContent    *string `json:"text_content,omitempty"`
	MediaCloudName *string `json:"media_cloud_name,omitempty"`
	Duration       *int64  `json:"duration,omitempty"`
	Caption        *string `json:"caption,omitempty"`
	Name           *string `json:"name,omitempty"`

	// fields to set when sending to client
	// Url            *string `json:"url,omitempty"`
	// MimeType       *string `json:"mime_type,omitempty"`
	// Size           *int64  `json:"size,omitempty"`
	// Extension      *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type     string `json:"type"`
	MsgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	return validation.ValidateStruct(&m,
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
}

type sendChatMsgAcd struct {
	PartnerUsername  string     `json:"toUser"`
	IsReply          bool       `json:"isReply"`
	ReplyTargetMsgId string     `json:"replyTargetMsgId"`
	Msg              MsgContent `json:"msg"`
	At               int64      `json:"at"`
}

func (vb sendChatMsgAcd) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required),
	)

	if err != nil {
		return helpers.ValidationError(err, "rcValidation.go", "sendChatMsgAcd")
	}

	/* validate and (if needed) clean bad uploaded media */
	if mediaCloudName := vb.Msg.MediaCloudName; mediaCloudName != nil {
		go func(msgType, mediaCloudName string) {

			ctx := context.Background()

			switch msgType {
			case "photo", "video":
				var (
					mcnBlur   string
					mcnActual string
				)

				fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &mcnBlur, &mcnActual)

				if mInfo := gcsHelpers.GetMediaInfo(ctx, mcnBlur); mInfo != nil {
					if mInfo.Size < 1*1024 || mInfo.Size > 100*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mcnBlur)
					}
				}

				if mInfo := gcsHelpers.GetMediaInfo(ctx, mcnActual); mInfo != nil {
					if msgType == "photo" && mInfo.Size < 1*1024 || mInfo.Size > 10*1024*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mcnActual)
					} else if mInfo.Size < 1*1024 || mInfo.Size > 40*1024*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mcnActual)
					}
				}
			case "voice":
				if mInfo := gcsHelpers.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 10*1024*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			case "audio":
				if mInfo := gcsHelpers.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 20*1024*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			default:
				if mInfo := gcsHelpers.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 50*1024*1024 {
						gcsHelpers.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			}
		}(vb.Msg.Type, *mediaCloudName)
	}

	return nil
}

type ackChatMsgDeliveredAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d ackChatMsgDeliveredAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "ackChatMsgDeliveredAcd")
}

type ackChatMsgReadAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d ackChatMsgReadAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "ackChatMsgReadAcd")
}

type reactToChatMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	Emoji           string `json:"emoji"`
	At              int64  `json:"at"`
}

func (d reactToChatMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.Emoji, validation.Required, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "reactToChatMsgAcd")
}

type removeReactionToChatMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	At              int64  `json:"at"`
}

func (d removeReactionToChatMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.At, validation.Required),
	)

	return helpers.ValidationError(err, "rcValidation.go", "removeReactionToChatMsgAcd")
}

type deleteChatMsgAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	MsgId           string `json:"msgId"`
	DeleteFor       string `json:"deleteFor"`
}

func (d deleteChatMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.DeleteFor, validation.Required, validation.In("me", "everyone").Error("expected value: 'me' or 'everyone'. but found "+d.DeleteFor)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "deleteChatMsgAcd")
}
