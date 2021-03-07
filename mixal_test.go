package main

/*
import (
	"testing"
)

var a = NewAssembler()

func TestAtom(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{Line: "*", Want: 0},
		{Line: "12345", Want: 12345},
		{Line: "somesym", Want: 0}, // need to define symbols
	}
	for _, test := range tests {
		v, _ := a.atom(test.Line)
		if v != test.Want {
			t.Error(test.Want.view(), v.view())
		}
	}
}

func TestExpression(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{Line: "-12345", Want: -12345},
		{Line: "123+45", Want: 168},
		{Line: "1:5", Want: 13},
	}
	for _, test := range tests {
		v, _ := a.expression(test.Line)
		if v != test.Want {
			t.Error(test.Want.view(), v.view())
		}
	}
}

func TestF(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{Line: "", Want: 5},
		{Line: "(1:5)", Want: 13},
	}
	for _, test := range tests {
		v, _ := a.f(test.Line)
		if v != test.Want {
			t.Error(test.Want.view(), v.view())
		}
	}
}

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
		if v, _ := a.wValue(test.Line); v != test.Want {
			t.Error(test.Want.view(), v.view())
		}
	}
}*/
