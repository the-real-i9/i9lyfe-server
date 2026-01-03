package appTypes

import (
	"github.com/goccy/go-json"
)

type ClientUser struct {
	Username string `json:"username"`
	Name     string `json:"name"`
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
