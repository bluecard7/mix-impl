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
	if w.value(0, 5) != -0xBEBEEFD {
		t.Error("unexpected value")
	}
}

//func TestAdd

func TestBitslice(t *testing.T) {
	var w Word = 1<<30 | 0x1234567
	s1, s2 := w.slice(1, 2), w.slice(3, 5) // have s1.start == 0
	//if s1.sign() != 1 << 31 || s2.sign() != 0 {
	//	t.Error("unexpected signs")
	//}
	bitcopy(s1, s2)
	if s1.word != 0x159F4567 {
		t.Errorf("0x%X != 0x%X", s1.word, 0x159F4567)
	}
}
