package chatMessageTypes

import (
	"i9lyfe/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Text struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Content string `json:"content"`
	} `json:"props"`
}

func (m Text) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Content, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "Text")
}

type Voice struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Duration int64  `json:"duration"`
		Data     []byte `json:"data"`
	} `json:"props"`
}

func (m Voice) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&m.Props.Data, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "Voice")
}

type Photo struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Data    []byte `json:"data"`
		Size    int64  `json:"size"`
		Summary string `json:"summary"`
	} `json:"props"`
}

func (m Photo) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Data, validation.Required),
		validation.Field(&m.Props.Size, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "Photo")
}

type Video struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Duration int64  `json:"duration"`
		Data     []byte `json:"data"`
		Size     int64  `json:"size"`
		Summary  string `json:"summary"`
	} `json:"props"`
}

func (m Video) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&m.Props.Data, validation.Required),
		validation.Field(&m.Props.Size, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "Video")
}

type Audio struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Name     string `json:"name"`
		Duration int64  `json:"duration"`
		Data     []byte `json:"data"`
		Size     int64  `json:"size"`
	} `json:"props"`
}

func (m Audio) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Name, validation.Required),
		validation.Field(&m.Props.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&m.Props.Data, validation.Required),
		validation.Field(&m.Props.Size, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "Audio")
}

type File struct {
	To    string `json:"to"`
	At    int64  `json:"at"`
	Props struct {
		Name string `json:"name"`
		Data []byte `json:"data"`
		Size int64  `json:"size"`
		Ext  string `json:"ext"`
	} `json:"props"`
}

func (m File) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.To, validation.Required),
		validation.Field(&m.At, validation.Required),
		validation.Field(&m.Props.Name, validation.Required),
		validation.Field(&m.Props.Data, validation.Required),
		validation.Field(&m.Props.Size, validation.Required),
		validation.Field(&m.Props.Ext, validation.Required),
	)

	return helpers.ValidationError(err, "msgTypes.go", "File")
}
