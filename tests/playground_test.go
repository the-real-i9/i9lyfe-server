package tests

import (
	"fmt"
	"testing"
)

func XTestPlayground(t *testing.T) {
	str := "small:smalling medium:mediuming large:larging"

	var (
		small  string
		medium string
		large  string
	)
	_, err := fmt.Sscanf(str, "small:%s medium:%s large:%s", &small, &medium, &large)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(small, medium, large)
}
