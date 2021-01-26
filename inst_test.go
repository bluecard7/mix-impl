package main

import (
	"testing"
)

func TestLDA(t *testing.T) {
	tests := []struct {
		Content MIXWord
		Inst    string
		Want    MIXWord // one for result, another for machine-lvl instruction
	}{
		{
			Content: NewWord(NEG_SIGN, 8, 60, 45),
			Inst:    "LDA 2000",
		},
		// NewWord with: list of size > 6, values greater than 63
	}
	machine := NewMachine()
	for _, test := range tests {
		// later inst, err
		inst := parseInst(test.Inst)
		machine.WriteCell(test.CellNum, test.Content)
		inst.Exec(machine, inst)
		// Read from cell and check if expected fields were loaded into rA
		t.Error(machine.Regs.A)
	}
}
