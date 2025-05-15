package chatMessageTypes

import (
	"fmt"
	"i9lyfe/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type textProps struct {
	Content string `json:"content"`
}

func (mp textProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Content, validation.Required),
	)
}

type Text struct {
	Props  textProps `json:"props"`
	ToUser string    `json:"toUser"`
	At     int64     `json:"at"`
}

func (m Text) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
		validation.Field(&m.ToUser, validation.Required),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "Text")
}

type voiceProps struct {
	Duration int64  `json:"duration"`
	Data     []byte `json:"data"`
}

func (mp voiceProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&mp.Data, validation.Required),
	)
}

type Voice struct {
	Props  voiceProps `json:"props"`
	ToUser string     `json:"toUser"`
	At     int64      `json:"at"`
}

func (m Voice) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.ToUser, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "Voice")
}

type photoProps struct {
	Data    []byte `json:"data"`
	Size    int64  `json:"size"`
	Caption string `json:"caption"`
}

func (mp photoProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Data, validation.Required),
		validation.Field(&mp.Size, validation.Required, validation.By(func(value any) error {
			if value.(int64) != int64(len(mp.Data)) {
				return fmt.Errorf("size specified does not match the calculated data size")
			}

			return nil
		})),
	)
}

type Photo struct {
	Props  photoProps `json:"props"`
	ToUser string     `json:"toUser"`
	At     int64      `json:"at"`
}

func (m Photo) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.ToUser, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "Photo")
}

type videoProps struct {
	Duration int64  `json:"duration"`
	Data     []byte `json:"data"`
	Size     int64  `json:"size"`
	Caption  string `json:"caption"`
}

func (mp videoProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&mp.Data, validation.Required, validation.By(func(value any) error {
			if value.(int64) != int64(len(mp.Data)) {
				return fmt.Errorf("size specified does not match the calculated data size")
			}

			return nil
		})),
		validation.Field(&mp.Size, validation.Required),
	)
}

type Video struct {
	Props  videoProps `json:"props"`
	ToUser string     `json:"toUser"`
	At     int64      `json:"at"`
}

func (m Video) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.ToUser, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "Video")
}

type audioProps struct {
	Name     string `json:"name"`
	Duration int64  `json:"duration"`
	Data     []byte `json:"data"`
	Size     int64  `json:"size"`
}

func (mp audioProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Name, validation.Required),
		validation.Field(&mp.Duration, validation.Required, validation.Min(1000).Error("duration can't be less than 1000msec")),
		validation.Field(&mp.Data, validation.Required),
		validation.Field(&mp.Size, validation.Required, validation.By(func(value any) error {
			if value.(int64) != int64(len(mp.Data)) {
				return fmt.Errorf("size specified does not match the calculated data size")
			}

			return nil
		})),
	)
}

type Audio struct {
	Props  audioProps `json:"props"`
	ToUser string     `json:"toUser"`
	At     int64      `json:"at"`
}

func (m Audio) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.ToUser, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "Audio")
}

type fileProps struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

func (mp fileProps) Validate() error {
	return validation.ValidateStruct(&mp,
		validation.Field(&mp.Name, validation.Required),
		validation.Field(&mp.Data, validation.Required),
		validation.Field(&mp.Size, validation.Required, validation.By(func(value any) error {
			if value.(int64) != int64(len(mp.Data)) {
				return fmt.Errorf("size specified does not match the calculated data size")
			}

			return nil
		})),
		validation.Field(&mp.Ext, validation.Required),
	)
}

type File struct {
	Props  fileProps `json:"props"`
	ToUser string    `json:"toUser"`
	At     int64     `json:"at"`
}

func (m File) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.ToUser, validation.Required),
		validation.Field(&m.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "chatMessageTypes.go", "File")
}
