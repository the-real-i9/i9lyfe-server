package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestPlayground(t *testing.T) {
	mp := &map[string]any{}

	jsonByte := bytes.NewReader([]byte(`{"e": 2, "d": "f"}`))

	if err := json.NewDecoder(jsonByte).Decode(mp); err != nil {
		t.Error(err)
	}

	fmt.Println((*mp)["d"])
}
