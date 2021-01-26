package main

var instTemplates = map[string]func() *Instruction{
	"LDA": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 8),
			Exec: func(machine *MIXArch, inst *Instruction) {
				contents := machine.ReadCell(inst.A())
				L, R := inst.F()
				machine.Regs.A[0] = POS_SIGN
				if L == 0 {
					machine.Regs.A[0] = contents[0]
					L = 1
				}
				partial := contents[L : R+1]
				copy(machine.Regs.A[6-len(partial):], partial)
			},
		}
	},
}
