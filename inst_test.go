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
			Content: NewWord(NEG_SIGN, 8, 60, 4, 5),
			Inst:    "LDA 2000",
		},
		{
			Content: NewWord(NEG_SIGN, 8, 60, 4, 5),
			Inst:    "LDA 2000(0:3)",
		},
		// NewWord with: list of size > 6, values greater than 63
	}
	machine := NewMachine()
	for _, test := range tests {
		// later inst, err
		inst := parseInst(test.Inst)
		copy(machine.Regs.A, NewWord())
		machine.WriteCell(inst.A(), test.Content)
		inst.Exec(machine, inst)
		t.Error(machine.Regs.A)
	}
}
