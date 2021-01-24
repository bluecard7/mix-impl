package main

/*
Involve timing
*/

type Instruction struct {
	C    BinByte
	Time int // time it takes to execute instruction
	Exec func(rSet RegisterSet, memory []BinWord)
}

var insts = map[string]Instruction{
	"LDA": Instruction{
		C: 8,
		// integral types used because program is parsed
		Exec: func(arch MIXArch, cell uint16, L, R byte) {
			contents := arch.Mem[cell]
			arch.Regs.A.Bytes[0] = 1
			if L == 0 {
				arch.Regs.A.Bytes[0] = contents[0]
				L = 1
			}
			partial := contents[L : R+1]
			copy(arch.Regs.A.Bytes[6-len(partial):], partial)
		},
	},
}
