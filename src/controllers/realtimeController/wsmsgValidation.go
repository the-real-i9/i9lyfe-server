package realtimeController

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"
	"slices"
	"strings"

	"cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
)

type rtActionBody struct {
	Action string `json:"action"`
	Data   any    `json:"data"`
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "rtActionBody")
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "getChatHistoryAcd")
}

type subToUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb subToUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "subToUserPresenceAcd")
}

type unsubFromUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb unsubFromUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "unsubFromUserPresenceAcd")
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
				validation.By(func(value any) error {
					val := value.(string)

					if !(strings.HasPrefix(val, "blur_placeholder:uploads/chat/") && strings.Contains(val, " actual:uploads/chat/")) {
						return errors.New("invalid media cloud name")
					}

					return nil
				}),
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
		return helpers.ValidationError(err, "realtimeController_validation.go", "sendChatMsgAcd")
	}

	if mediaCloudName := *vb.Msg.MediaCloudName; mediaCloudName != "" {
		_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mediaCloudName).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: media (%s) does not exist in cloud", mediaCloudName))
		}
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "ackChatMsgDeliveredAcd")
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "ackChatMsgReadAcd")
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "reactToChatMsgAcd")
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "removeReactionToChatMsgAcd")
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

	return helpers.ValidationError(err, "realtimeController_validation.go", "deleteChatMsgAcd")
}
