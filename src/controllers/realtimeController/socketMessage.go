package realtimeController

import (
	"i9lyfe/src/helpers"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type clientMessageBody struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func (b clientMessageBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Event, validation.Required, validation.In(
			"send message",
			"get chat history",
			"ack message delivered",
			"ack message read",
			"react to message",
			"remove reaction to message",
			"delete message",
			"start receiving post updates",
			"stop receiving post updates",
		)),
		validation.Field(&b.Data,
			validation.Required.When(
				!slices.Contains([]string{"start receiving post updates", "stop receiving post updates"}, b.Event),
			),
		),
	)

	return helpers.ValidationError(err, "realtimeController_socketMessage.go", "clientEvclientMessageBodyentBody")
}
