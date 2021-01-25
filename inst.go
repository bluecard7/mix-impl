package main

// May be better to make Instruction an interface
// because how would you handle binary ops in this arrangement?
type Instruction struct {
	C    MIXByte
	Time int // time it takes to execute instruction
	Exec func(arch MIXArch, cell uint16, L, R byte)
}

var insts = map[string]Instruction{
	"LDA": Instruction{
		C: 8,
		// integral types used because program is parsed
		Exec: func(arch MIXArch, cell uint16, L, R byte) {
			contents := arch.Mem[cell]
			arch.Regs.A[0] = 1
			if L == 0 {
				arch.Regs.A[0] = contents[0]
				L = 1
			}
			partial := contents[L : R+1]
			copy(arch.Regs.A[6-len(partial):], partial)
		},
	},
}
