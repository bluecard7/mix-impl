package main

import (
	"testing"
)

func TestLDA(t *testing.T) {
	tests := []struct {
		Content MIXWord
		CellNum uint16
		L, R    byte
		Want    MIXWord
	}{
		{
			Content: NewWord(NEGATIVE, 8, 60, 45), // need to allow initializing of value
			CellNum: 2000,
			L:       0,
			R:       5,
		},
		// NewWord with: list of size > 6, values greater than 63
	}
	machine := NewMachine()
	for _, test := range tests {
		inst := insts["LDA"]
		machine.WriteCell(test.CellNum, test.Content)
		inst.Exec(machine, test.CellNum, test.L, test.R)
		// Read from cell and check if expected fields were loaded into rA
		t.Error(machine.ReadCell(test.CellNum))
	}
}
