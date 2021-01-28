package main

import (
	"testing"
)

// test parseInst separately? would test code
func TestParseInst(t *testing.T) {
	tests := []struct {
		Line string
		Code MIXWord
	}{}

	for _, test := range tests {
		t.Error(test)
	}
}

// TestLD tests load and load negative instructions.
func TestLD(t *testing.T) {
	tests := []struct {
		Content      MIXWord
		Line         string
		RegI         int
		Code, Result MIXWord
	}{
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000",
			RegI:    A,
			Code:    NewWord(POS_SIGN, 31, 16, 0, 5, 8),
			Result:  NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
		},
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000(0:3)",
			RegI:    A,
			Code:    NewWord(POS_SIGN, 31, 16, 0, 3, 8),
			Result:  NewWord(NEG_SIGN, 0, 0, 1, 2, 3),
		},
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000,4(4:5)",
			RegI:    A,
			Code:    NewWord(POS_SIGN, 31, 16, 4, 37, 8),
			Result:  NewWord(POS_SIGN, 0, 0, 0, 4, 5),
		},
		// NewWord with: list of size > 6, values greater than 63
	}
	machine := NewMachine()
	for _, test := range tests {
		// later inst, err
		inst, _ := ParseInst(test.Line)
		copy(machine.R[test.RegI], NewWord()) // resets register
		machine.WriteCell(inst.A(), test.Content)
		inst.Exec(machine, inst)
		if !inst.Code.Raw().Equals(test.Code.Raw()) {
			t.Errorf("Codes don't match for %s: want %v, got %v", test.Line, test.Code, inst.Code)
		}
		if !machine.R[test.RegI].Raw().Equals(test.Result.Raw()) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Result, machine.R[test.RegI])
		}
	}
}

// TestST tests store instructions.
func TestST(t *testing.T) {
}
