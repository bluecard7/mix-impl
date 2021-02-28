package main

import (
	"testing"
)

func TestWordBasics(t *testing.T) {
	var w Word = 0x4BEBEEFD
	if w.sign() != 1<<30 {
		t.Error("unexpected sign")
	}
	if w.data() != 0xBEBEEFD { // 32nd bit not used
		t.Error("unexpected data")
	}
}

//func TestAdd

func TestBitslice(t *testing.T) {
	var w Word = 1<<30 | 0x1234567
	if w.slice(0, 5).value() != -0x1234567 {
		t.Error("unexpected value")
	}
	s1, s2 := w.slice(0, 2), w.slice(3, 5) // have s1.start == 0
	s1.copy(s2)
	if s1.word != 0x159F4567 {
		t.Errorf("0x%X != 0x%X", s1.word, 0x159F4567)
	}
}
