package appTypes

import (
	"encoding/json"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ClientUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

func (c ClientUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableMap map[string]any

func (c BinableMap) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableSlice []string

func (c BinableSlice) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type ServerEventMsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
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
	Type      string `json:"type"`
	*MsgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	msgType := m.Type

	return validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "photo", "file").Error("invalid message type"),
		),
		validation.Field(&m.MediaCloudName,
			validation.When(msgType == "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required),
		),
		validation.Field(&m.MsgProps, validation.Required),
		validation.Field(&m.TextContent, validation.When(msgType != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Duration, validation.When(msgType != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file", "audio"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
	)
}
