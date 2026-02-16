package appTypes

import "github.com/vmihailenco/msgpack/v5"

type ClientUser struct {
	Username string `msgpack:"username"`
}

type BinableMap map[string]any

func (c BinableMap) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(c)
}

type BinableSlice []string

func (c BinableSlice) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(c)
}

type ServerEventMsg struct {
	Event string `msgpack:"event"`
	Data  any    `msgpack:"data"`
}
