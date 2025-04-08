package realtimeController

import (
	"i9lyfe/src/helpers"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type clientMessageBody struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func (b clientMessageBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Event, validation.Required, validation.In(
			"start receiving post updates",
			"stop receiving post updates",
			"chat: send message: text",
			"chat: send message: voice",
			"chat: send message: photo",
			"chat: send message: video",
			"chat: send message: audio",
			"chat: send message: file",
			"chat: get history",
			"chat: ack message delivered",
			"chat: ack message read",
			"chat: react to message",
			"chat: remove reaction to message",
			"chat: delete message",
		)),
		validation.Field(&b.Data,
			validation.Required.When(
				!slices.Contains([]string{"start receiving post updates", "stop receiving post updates"}, b.Event),
			),
		),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "clientMessageBody")
}

type getChatHistoryEvd struct {
	PartnerUsername string `json:"partner_username"`
	Offset          int64  `json:"offset"`
}

func (b getChatHistoryEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "getChatHistoryEvd")
}

type ackChatMsgDeliveredEvd struct {
	PartnerUsername string `json:"partner_username"`
	MsgId           string `json:"message_id"`
	At              int64  `json:"at"`
}

func (b ackChatMsgDeliveredEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.MsgId, validation.Required),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "ackChatMsgDeliveredEvd")
}

type ackChatMsgReadEvd struct {
	PartnerUsername string `json:"partner_username"`
	MsgId           string `json:"message_id"`
	At              int64  `json:"at"`
}

func (b ackChatMsgReadEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.MsgId, validation.Required),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "ackChatMsgReadEvd")
}

type reactToChatMsgEvd struct {
	PartnerUsername string `json:"partner_username"`
	MsgId           string `json:"message_id"`
	Reaction        string `json:"reaction"`
	At              int64  `json:"at"`
}

func (b reactToChatMsgEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.MsgId, validation.Required),
		validation.Field(&b.Reaction, validation.Required, validation.Required, validation.RuneLength(1, 1).Error("expected an emoji character"), is.Multibyte.Error("expected an emoji character")),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "reactToChatMsgEvd")
}

type removeReactionToChatMsgEvd struct {
	PartnerUsername string `json:"partner_username"`
	MsgId           string `json:"message_id"`
	At              int64  `json:"at"`
}

func (b removeReactionToChatMsgEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.MsgId, validation.Required),
		validation.Field(&b.At, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "removeReactionToChatMsgEvd")
}

type deleteChatMsgEvd struct {
	PartnerUsername string `json:"partner_username"`
	MsgId           string `json:"message_id"`
	For             string `json:"for"`
}

func (b deleteChatMsgEvd) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.MsgId, validation.Required),
		validation.Field(&b.For, validation.Required, validation.In("me", "everyone")),
	)

	return helpers.ValidationError(err, "realtimeController_clientSockMsg.go", "deleteChatMsgEvd")
}
