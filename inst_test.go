package main

import (
	"errors"
	"testing"
)

var machine = NewMachine()

func TestInst(t *testing.T) {
	var inst Word = 0x41234567
	if inst.a() != -0x48 || inst.i() != 0x34 || inst.f() != 0x54 || inst.c() != 0x27 {
		t.Error(inst.a(), inst.i(), inst.f(), inst.c())
	}
}

/*
func TestParseInst(t *testing.T) {
	tests := []struct {
		Line   string
		Err    error
		Fields MIXBytes
	}{
		{Line: "LDA 2000", Fields: MIXBytes{POS_SIGN, 31, 16, 0, 5, 8}},        // basic
		{Line: "LDXN 2000", Fields: MIXBytes{POS_SIGN, 31, 16, 0, 5, 23}},      // different op len
		{Line: "LDA 27", Fields: MIXBytes{POS_SIGN, 0, 27, 0, 5, 8}},           // different address len
		{Line: "LDA -345", Fields: MIXBytes{NEG_SIGN, 5, 25, 0, 5, 8}},         // negative address
		{Line: "LDA 2000,5", Fields: MIXBytes{POS_SIGN, 31, 16, 5, 5, 8}},      // use index
		{Line: "LDA 2000(1:4)", Fields: MIXBytes{POS_SIGN, 31, 16, 0, 12, 8}},  // use field spec
		{Line: "LDA 2000,3(0:0)", Fields: MIXBytes{POS_SIGN, 31, 16, 3, 0, 8}}, // use both
		{Line: "LDA9 2AA,9I(.3:-4)", Err: ErrRegex},                            // regex failure
		{Line: "HELLO 30416", Err: ErrOp},                                      // undefined op
		{Line: "LDA 2000,7", Err: ErrIndex},                                    // out of bound index
		{Line: "LDA 2000(7:5)", Err: ErrField},                                 // out of bound L
		{Line: "LDA 2000(2:7)", Err: ErrField},                                 // out of bound R
		{Line: "LDA 2000(3:2)", Err: ErrField},                                 // L > R
	}

	for _, test := range tests {
		inst, err := ParseInst(test.Line)
		if test.Err != nil && !errors.Is(test.Err, err) {
			t.Errorf("Errors don't match for \"%s\": want \"%v\", got \"%v\"", test.Line, test.Err, err)
		}
		if test.Fields != nil && !test.Fields.Equals(inst.Fields()) {
			t.Errorf("Fields don't match for \"%s\": want \"%v\", got \"%v\"", test.Line, test.Fields, inst.Fields())
		}
	}
}
*/

// TestLD tests load and load negative instructions.
func TestLD(t *testing.T) {
	tests := []struct {
		Inst, Want Word
		RegI       int
	}{
		{
			Inst: composeInst(2000, 0, 5, C_LD), // LDA 2000
			RegI: A,
			//Want: 0x41083105, // NEG_SIGN, 1, 2, 3, 4, 5
			Want: composeWord(1, 1, 2, 3, 4, 5),
		},
		{
			//Line: "LDA 2000(0:3)",
			Inst: composeInst(2000, 0, 3, C_LD), // LDA 2000(0:3)
			RegI: A,
			//Want: 0x40001083, // NEG_SIGN, 0, 0, 1, 2, 3
			Want: composeWord(1, 0, 0, 1, 2, 3),
		},
		{
			Inst: composeInst(2000, 4, 37, C_LD), // LDA 2000,4(4:5), but index has no effect here
			RegI: A,
			//Want: 0x00000105, // POS_SIGN, 0, 0, 0, 4, 5
			Want: composeWord(0, 0, 0, 0, 4, 5),
		},
		{
			//Line: "LD1 2000", // ignores bytes 1-3, book says undefined if set to nonzero #
			Inst: composeInst(2000, 0, 5, C_LD+I1),
			RegI: I1,
			//Want: 0x50140000, // NEG_SIGN, 4, 5
			Want: composeWord(1, 4, 5, 0, 0, 0),
		},
		{
			Inst: composeInst(2000, 0, 5, C_LDN),
			RegI: A,
			Want: composeWord(0, 1, 2, 3, 4, 5),
		},
	}
	for _, test := range tests {
		/*inst, err := ParseInst(test.Line)
		if err != nil {
			t.Errorf("Error parsing %s: %v", test.Line, err)
		}*/
		// Want a way to generate insts with A, I, F, C
		copy(machine.R[test.RegI], 0)        // resets register
		copy(machine.Cell(inst), 0x41083105) // default cell
		machine.Exec(inst)
		result := machine.R[test.RegI].Raw()
		if !test.Want.Equals(result) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Want, result)
		}
	}
}

/*
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
		copy(machine.Cell(inst), MIXBytes{NEG_SIGN, 1, 2, 3, 4, 5}) // default cell
		machine.Exec(inst)
		result := machine.Cell(inst)
		if !test.Want.Equals(result) {
			t.Errorf("Incorrect result for %s: want %v, got %v", test.Line, test.Want, result)
		}
	}
}

/*
func TestArithmetic(t *testing.T) {
	tests := []struct {
		Line         string
		RAData, Want MIXBytes
	}{
		{
			Line:   "ADD 1000",
			RAData: MIXBytes{},
			Want:   MIXBytes{},
		},
	}
	for _, test := range tests {
		_, err := ParseInst(test.Line)
		if err != nil {
			t.Fatalf("Error parsing %s: %v", test.Line, err)
		}
		t.Error("N/A")
	}
}

func TestAddressTransfer(t *testing.T) {
	t.Error("N/A")
}

func TestComparison(t *testing.T) {
	t.Error("N/A")
}

func TestJump(t *testing.T) {
	t.Error("N/A")
}

func TestShift(t *testing.T) {
	t.Error("N/A")
}

func TestMove(t *testing.T) {
	t.Error("N/A")
}

func TestNop(t *testing.T) {
	t.Error("N/A")
}

func TestHalt(t *testing.T) {
	t.Error("N/A")
}

func TestIO(t *testing.T) {
	t.Error("N/A")
}

func TestConversion(t *testing.T) {
	t.Error("N/A")
}
*/
