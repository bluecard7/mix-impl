package main

import (
	"testing"
)

func TestWordBasics(t *testing.T) {
	w := -composeWord(1, 2, 3, 4, 5)
	if w.sign() != -1 || w.data() != -w {
		t.Error("unexpected word")
	}

	sliced, want := w.slice(0, 2).w, -composeWord(0, 0, 0, 1, 2)
	if want != sliced {
		t.Errorf("\nWant:%s\nGot:%s\n", want.view(), sliced.view())
	}
}

func TestBitslice(t *testing.T) {
	var w Word = -composeWord(1, 2, 3, 4, 5)
	s1, s2 := w.slice(0, 2), w.slice(3, 5)
	s1.copy(s2)
	if want := -composeWord(0, 0, 0, 4, 5); want != s1.w {
		t.Errorf("\nWant:%s\nGot:%s\n", want.view(), s1.w.view())
	}
}
