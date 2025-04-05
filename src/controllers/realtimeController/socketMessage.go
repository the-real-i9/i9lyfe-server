package realtimeController

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type clientMessageBody struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func (b clientMessageBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Event, validation.Required),
		validation.Field(&b.Data, validation.Required.When(true /* come back here */)),
	)

	return helpers.ValidationError(err, "realtimeController_socketMessage.go", "clientEvclientMessageBodyentBody")
}
