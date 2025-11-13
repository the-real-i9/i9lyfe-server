package tests

import (
	"i9lyfe/src/helpers"
	"testing"
)

func XTestPlayground(t *testing.T) {
	type udt struct {
		A int    `json:"a"`
		B string `json:"b"`
	}

	mp := []map[string]any{{"a": 1, "b": "c"}}

	var lett []udt

	helpers.ToStruct(mp, &lett)

	t.Logf("%+v", lett)
}
