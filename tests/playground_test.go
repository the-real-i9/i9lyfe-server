package tests

import (
	"sync"
	"testing"
)

func TestPlayground(t *testing.T) {
	type samp struct {
		A int
		B string
	}

	mp := samp{A: 1, B: "c"}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(mymp samp) {
		defer wg.Done()

		mymp.A = 2
		t.Log(mp, mymp)
	}(mp)

	go func(mymp samp) {
		defer wg.Done()

		mymp.B = "d"
		t.Log(mp, mymp)
	}(mp)

	wg.Wait()

	t.Log(mp)
}
