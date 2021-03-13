package main

import (
	"testing"
)

var a = NewAssembler()

func wordDiff(want, got Word) string {
	return "\nWant:" + want.view() + "\nGot:" + got.view()
}

func TestAtom(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"*", 0, nil},
		{"12345", 12345, nil},
		{"SOMESYM", 0, ErrNonAtom}, // need to define symbols
	}
	for _, test := range tests {
		v, err := a.atom(test.Line)
		if v != test.Want || err != test.Err {
			t.Error(test.Line)
			t.Error(wordDiff(test.Want, v))
			t.Error(test.Err, "|", err)
		}
	}
}

func TestExpression(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"-12345", -12345, nil},
		{"123+45", 168, nil},
		{"1:5", 13, nil},
	}
	for _, test := range tests {
		v, err := a.expression(test.Line)
		if v != test.Want || err != test.Err {
			t.Error(wordDiff(test.Want, v))
		}
	}
}

func TestLiteral(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"=1+34(5+4)=", composeWord(35, 0, 0, 0, 0), nil},
	}
	for _, test := range tests {
		v, err := a.literal(test.Line)
		if v != test.Want || err != test.Err {
			t.Error(wordDiff(test.Want, v))
		}
	}
}

func TestA(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"", 0, nil}, // just vacuous case since it's a proxy to literal and expression
		// futureRef test?
	}
	for _, test := range tests {
		v, err := a.a(test.Line)
		if v != test.Want || err != test.Err {
			t.Error(test.Line)
			t.Error(wordDiff(test.Want, v))
			t.Error(test.Err, "|", err)
		}
	}
}

func TestI(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"", 0, nil},
		{",1:5", 13, nil},
	}
	for _, test := range tests {
		v, err := a.i(test.Line)
		if v != test.Want || err != test.Err {
			t.Error(wordDiff(test.Want, v))
		}
	}
}

func TestF(t *testing.T) {
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"", 5, nil},
		{"(1:5)", 13, nil},
	}
	for _, test := range tests {
		v, err := a.f(test.Line)
		if v != test.Want || err != test.Err {
			t.Errorf("\nWant:%s\nGot:%s\n", test.Want.view(), v.view())
		}
	}
}

func TestWValue(t *testing.T) {
	// Test once (w/ or w/o F)
	tests := []struct {
		Line string
		Want Word
		Err  error
	}{
		{"-1000(0:2)", -composeWord(1000>>6&63, 1000&63, 0, 0, 0), nil},
		{"-1000", -composeWord(0, 0, 0, 1000>>6&63, 1000&63), nil},
		{"-1000(0:2),1", composeWord(0, 0, 0, 0, 1), nil},
		{"1,-1000(0:2)", -composeWord(1000>>6&63, 1000&63, 0, 0, 1), nil},
	}

	for _, test := range tests {
		v, err := a.wValue(test.Line)
		if v != test.Want || err != test.Err {
			t.Errorf("\nWant:%s\nGot:%s\n", test.Want.view(), v.view())
			t.Error(err)
		}
	}
}
