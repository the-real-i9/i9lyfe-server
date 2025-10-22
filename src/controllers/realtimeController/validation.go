package realtimeController

import (
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"slices"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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
					"start receiving post updates",
					"stop receiving post updates",
					"start receiving comment updates",
					"stop receiving comment updates",
				}, b.Action),
			),
		),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "rtActionBody")
}

type getChatHistoryAcd struct {
	PartnerUsername string `json:"partnerUsername"`
	Offset          int64  `json:"offset"`
}

func (d getChatHistoryAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "getChatHistoryAcd")
}

type sendChatMsgAcd struct {
	PartnerUsername  string               `json:"partnerUsername"`
	IsReply          bool                 `json:"isReply"`
	ReplyTargetMsgId string               `json:"replyTargetMsgId"`
	Msg              *appTypes.MsgContent `json:"msg"`
	At               int64                `json:"at"`
}

func (vb sendChatMsgAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "sendChatMsgAcd")
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
	Reaction        string `json:"reaction"`
	At              int64  `json:"at"`
}

func (d reactToChatMsgAcd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.MsgId, validation.Required),
		validation.Field(&d.Reaction, validation.Required, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
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
		validation.Field(&d.DeleteFor, validation.Required, validation.In("me", "everyone", "").Error("expected value: 'me' or 'everyone', or no value. but found "+d.DeleteFor)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "deleteChatMsgAcd")
}
