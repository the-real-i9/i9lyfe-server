package tests

import (
	"testing"
)

func XTestPlayground(t *testing.T) {
	x := make(map[string]int)

	x["a"]++
	x["a"]++
	x["b"]--

	t.Log(x)
}
