package main

import (
	"testing"
)

var asmr = NewAssembler()

func TestWValue(t *testing.T) {
	// Test once (w/ or w/o F)
	tests := []struct {
		Line string
		Want Word
	}{
		{"-1000(0:2)", -composeWord(1000>>6&63, 1000&63, 0, 0, 0)},
		{"-1000", -composeWord(0, 0, 0, 1000>>6&63, 1000&63)},
		//
	}

	for _, test := range tests {
		if v, err := asmr.wValue(test.Line); v != test.Want {
			t.Error(v.view(), test.Want.view())
		}
	}
}
