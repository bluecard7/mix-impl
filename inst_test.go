package main

import (
	"testing"
)

func TestLDA(t *testing.T) {
	tests := []struct {
		Content      MIXWord
		Inst         string
		Code, Result MIXWord
	}{
		{
			Content: NewWord(NEG_SIGN, 8, 60, 4, 5),
			Inst:    "LDA 2000",
			Code:    NewWord(POS_SIGN),
			Result:  NewWord(NEG_SIGN, 8, 60, 4, 5),
		},
		{
			Content: NewWord(NEG_SIGN, 8, 60, 4, 5),
			Inst:    "LDA 2000(0:3)",
			Code:    NewWord(),
			Result:  NewWord(NEG_SIGN, 0, 0, 8, 60, 4),
		},
		{
			Content: NewWord(POS_SIGN),
			Inst:    "LDA 2500(4:5)",
			Code:    NewWord(),
			Result:  NewWord(),
		},
		// NewWord with: list of size > 6, values greater than 63
	}
	machine := NewMachine()
	for _, test := range tests {
		// later inst, err
		inst := parseInst(test.Inst)
		copy(machine.R[A], NewWord())
		machine.WriteCell(inst.A(), test.Content)
		inst.Exec(machine, inst)
		t.Error(machine.R[A])
	}
}
