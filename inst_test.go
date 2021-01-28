package main

import (
	"errors"
	"testing"
)

var machine = NewMachine()

func TestParseInst(t *testing.T) {
	tests := []struct {
		Line string
		Err  error
		Code MIXBytes
	}{
		{Line: "LDA 2000", Code: MIXBytes{POS_SIGN, 31, 16, 0, 5, 8}},        // basic
		{Line: "LDXN 2000", Code: MIXBytes{POS_SIGN, 31, 16, 0, 5, 23}},      // different op len
		{Line: "LDA 27", Code: MIXBytes{POS_SIGN, 0, 27, 0, 5, 8}},           // different address len
		{Line: "LDA -345", Code: MIXBytes{NEG_SIGN, 5, 25, 0, 5, 8}},         // negative address
		{Line: "LDA 2000,5", Code: MIXBytes{POS_SIGN, 31, 16, 5, 5, 8}},      // use index
		{Line: "LDA 2000(1:4)", Code: MIXBytes{POS_SIGN, 31, 16, 0, 12, 8}},  // use field spec
		{Line: "LDA 2000,3(0:0)", Code: MIXBytes{POS_SIGN, 31, 16, 3, 0, 8}}, // use both
		{Line: "LDA9 2AA,9I(.3:-4)", Err: ErrRegex},                          // regex failure
		{Line: "HELLO 30416", Err: ErrOp},                                    // undefined op
		{Line: "LDA 2000,7", Err: ErrIndex},                                  // out of bound index
		{Line: "LDA 2000(7:5)", Err: ErrField},                               // out of bound L
		{Line: "LDA 2000(2:7)", Err: ErrField},                               // out of bound R
		{Line: "LDA 2000(3:2)", Err: ErrField},                               // L > R
	}

	for _, test := range tests {
		inst, err := ParseInst(test.Line)
		if test.Err != nil && !errors.Is(test.Err, err) {
			t.Errorf("Errors don't match for \"%s\": want \"%v\", got \"%v\"", test.Line, test.Err, err)
		}
		if test.Code != nil && !test.Code.Equals(inst.Code) {
			t.Errorf("Codes don't match for \"%s\": want \"%v\", got \"%v\"", test.Line, test.Code, inst.Code)
		}
	}
}

// TestLD tests load and load negative instructions.
func TestLD(t *testing.T) {
	tests := []struct {
		Content MIXBytes
		Line    string
		RegI    int
		Result  MIXBytes
	}{
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000",
			RegI:    A,
			Result:  MIXBytes{NEG_SIGN, 1, 2, 3, 4, 5},
		},
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000(0:3)",
			RegI:    A,
			Result:  MIXBytes{NEG_SIGN, 0, 0, 1, 2, 3},
		},
		{
			Content: NewWord(NEG_SIGN, 1, 2, 3, 4, 5),
			Line:    "LDA 2000,4(4:5)", // has no effect as of now since rI? are all zeros
			RegI:    A,
			Result:  MIXBytes{POS_SIGN, 0, 0, 0, 4, 5},
		},
		// NewWord with: list of size > 6, values greater than 63 -> separately test
	}
	for _, test := range tests {
		inst, _ := ParseInst(test.Line)
		copy(machine.R[test.RegI], NewWord()) // resets register
		machine.WriteCell(inst.A(), test.Content)
		inst.Exec(machine, inst)
		if !test.Result.Equals(machine.R[test.RegI].Raw()) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Result, machine.R[test.RegI])
		}
	}
}

// TestST tests store instructions.
func TestST(t *testing.T) {
	tests := []struct {
		Content MIXBytes
		Line    string
		RegI    int
	}{}
	for _, test := range tests {
		inst, _ := ParseInst(test.Line)
		copy(machine.R[test.RegI], MIXBytes{POS_SIGN, 6, 7, 8, 9, 0})
		machine.WriteCell(inst.A(), test.Content)
	}
}
