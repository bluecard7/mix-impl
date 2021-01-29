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
		Line string
		RegI int
		Want MIXBytes
	}{
		{
			Line: "LDA 2000",
			RegI: A,
			Want: MIXBytes{NEG_SIGN, 1, 2, 3, 4, 5},
		},
		{
			Line: "LDA 2000(0:3)",
			RegI: A,
			Want: MIXBytes{NEG_SIGN, 0, 0, 1, 2, 3},
		},
		{
			Line: "LDA 2000,4(4:5)", // has no effect as of now since rI_ are all zeros
			RegI: A,
			Want: MIXBytes{POS_SIGN, 0, 0, 0, 4, 5},
		},
		// NewWord with: list of size > 6, values greater than 63 -> separately test
		{
			Line: "LD1 2000", // ignores bytes 1-3, book says undefined if set to nonzero #
			RegI: I1,
			Want: MIXBytes{NEG_SIGN, 4, 5},
		},
		{
			Line: "LDAN 2000",
			RegI: A,
			Want: MIXBytes{POS_SIGN, 1, 2, 3, 4, 5},
		},
	}
	for _, test := range tests {
		inst, err := ParseInst(test.Line)
		if err != nil {
			t.Errorf("Error parsing %s: %v", test.Line, err)
		}
		copy(machine.R[test.RegI], NewWord())                          // resets register
		machine.WriteCell(inst.A(), MIXBytes{NEG_SIGN, 1, 2, 3, 4, 5}) // default cell
		inst.Exec(machine, inst)
		result := machine.R[test.RegI].Raw()
		if !test.Want.Equals(result) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Want, result)
		}
	}
}

// TestST tests store instructions.
func TestST(t *testing.T) {
	tests := []struct {
		Line    string
		RegI    int
		RegData MIXBytes // data in register
		Want    MIXBytes // stored result in cell
	}{
		{
			Line:    "STA 2000",
			RegI:    A,
			RegData: MIXBytes{POS_SIGN, 6, 7, 8, 9, 0},
			Want:    MIXBytes{POS_SIGN, 6, 7, 8, 9, 0},
		},
		{
			Line:    "STA 2000(2:3)",
			RegI:    A,
			RegData: MIXBytes{POS_SIGN, 6, 7, 8, 9, 0},
			Want:    MIXBytes{NEG_SIGN, 1, 9, 0, 4, 5},
		},
		{
			Line:    "ST1 2000",
			RegI:    I1,
			RegData: MIXBytes{POS_SIGN, 9, 9},
			Want:    MIXBytes{POS_SIGN, 0, 0, 0, 9, 9},
		},
		{
			Line:    "ST1 2000(0:3)",
			RegI:    I1,
			RegData: MIXBytes{POS_SIGN, 9, 9},
			Want:    MIXBytes{POS_SIGN, 0, 9, 9, 4, 5},
		},
		{
			Line:    "STJ 2000",
			RegI:    J,
			RegData: MIXBytes{POS_SIGN, 9, 9},
			Want:    MIXBytes{POS_SIGN, 9, 9, 3, 4, 5},
		},
		{
			Line:    "STZ 2000",
			RegI:    I1,
			RegData: MIXBytes{POS_SIGN, 6, 7, 8, 9, 0},
			Want:    MIXBytes{POS_SIGN, 0, 0, 0, 0, 0},
		},
	}
	for _, test := range tests {
		inst, err := ParseInst(test.Line)
		if err != nil {
			t.Fatalf("Error parsing %s: %v", test.Line, err)
		}
		copy(machine.R[test.RegI], test.RegData)
		machine.WriteCell(inst.A(), MIXBytes{NEG_SIGN, 1, 2, 3, 4, 5}) // default cell
		inst.Exec(machine, inst)
		result := machine.ReadCell(inst.A())
		if !test.Want.Equals(result) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Want, result)
		}
	}
}
