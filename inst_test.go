package main

import (
	//	"errors"
	"testing"
)

var machine = NewMachine()

func TestInst(t *testing.T) {
	var inst Word = -composeInst(1000, 2, 3, 4)
	if inst.a() != -1000 || inst.i() != 2 || inst.f() != 3 || inst.c() != 4 {
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
			Want: -composeWord(1, 2, 3, 4, 5),
		},
		{
			Inst: composeInst(2000, 0, 3, C_LD), // LDA 2000(0:3)
			RegI: A,
			Want: -composeWord(0, 0, 1, 2, 3),
		},
		{
			Inst: composeInst(2000, 4, 37, C_LD), // LDA 2000,4(4:5), but index has no effect here
			RegI: A,
			Want: composeWord(0, 0, 0, 4, 5),
		},
		{
			//Line: "LD1 2000", // ignores bytes 1-3, book says undefined if set to nonzero #
			Inst: composeInst(2000, 0, 5, C_LD+I1),
			RegI: I1,
			Want: -composeWord(0, 0, 0, 4, 5),
		},
		{
			Inst: composeInst(2000, 0, 5, C_LDN),
			RegI: A,
			Want: composeWord(1, 2, 3, 4, 5),
		},
	}
	for _, test := range tests {
		inst := test.Inst
		machine.R[test.RegI].w = 0 // resets register
		machine.Write(inst.a(), -composeWord(1, 2, 3, 4, 5))
		machine.Exec(inst)
		result := machine.R[test.RegI].w
		if test.Want != result {
			t.Errorf("\n%s\n\nWant:%s\nGot%s\n", inst.instView(), test.Want.view(), result.view())
		}
	}
}

type RegState struct {
	I    int
	Data Word
}

// TestST tests store instructions.
func TestST(t *testing.T) {
	tests := []struct {
		Reg        RegState // to setup register
		Inst, Want Word
	}{
		{
			Inst: composeInst(2000, 0, 5, C_ST), // "STA 2000"
			Reg:  RegState{A, composeWord(6, 7, 8, 9, 0)},
			Want: composeWord(6, 7, 8, 9, 0),
		},
		{
			Inst: composeInst(2000, 0, 19, C_ST), // "STA 2000(2:3)"
			Reg:  RegState{A, composeWord(6, 7, 8, 9, 0)},
			Want: -composeWord(1, 9, 0, 4, 5),
		},
		{
			Inst: composeInst(2000, 0, 3, C_ST+I1), // "ST1 2000(0:3)"
			Reg:  RegState{I1, composeWord(0, 0, 0, 9, 9)},
			Want: composeWord(0, 9, 9, 4, 5),
		},
		{
			Inst: composeInst(2000, 0, 2, C_ST+J), // "STJ 2000"
			Reg:  RegState{J, composeWord(0, 0, 0, 9, 9)},
			Want: composeWord(9, 9, 3, 4, 5),
		},
		{
			Inst: composeInst(2000, 0, 5, 33), // "STZ 2000"
			Reg:  RegState{A, composeWord(6, 7, 8, 9, 0)},
			Want: composeWord(0, 0, 0, 0, 0),
		},
	}
	for _, test := range tests {
		inst := test.Inst
		machine.R[test.Reg.I].w = test.Reg.Data
		machine.Write(inst.a(), -composeWord(1, 2, 3, 4, 5)) // default cell
		machine.Exec(inst)
		if result := machine.Read(inst.a()); test.Want != result {
			t.Errorf("\n%s\n\nWant:%s\nGot:%s\n", inst.instView(), test.Want.view(), result.view())
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
