package main

type InstTemplates map[string]func() *Instruction

func (dst InstTemplates) merge(src InstTemplates) {
	for k, v := range src {
		dst[k] = v
	}
}

func LD(machine *MIXArch, inst *Instruction, register []MIXByte) {
	contents := machine.ReadCell(inst.A())
	L, R := inst.F()
	register[0] = POS_SIGN
	if L == 0 {
		register[0] = contents[0]
		L = 1
	}
	partial := contents[L : R+1]
	copy(register[6-len(partial):], partial)
}

var load = InstTemplates{
	"LDA": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 8),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.A)
			},
		}
	},
	"LD1": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 9),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I1)
			},
		}
	},
	"LD2": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 10),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I2)
			},
		}
	},
	"LD3": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 11),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I3)
			},
		}
	},
	"LD4": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 12),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I4)
			},
		}
	},
	"LD5": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 13),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I5)
			},
		}
	},
	"LD6": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 14),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.I6)
			},
		}
	},
	"LDX": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 15),
			Exec: func(machine *MIXArch, inst *Instruction) {
				LD(machine, inst, machine.Regs.X)
			},
		}
	},
	//"LDAN": nil, #16
	//"LDXN": nil, #23
	//"LDiN": nil, #17-22
}

func ST(machine *MIXArch, inst *Instruction, register []MIXByte) {
	L, R := inst.F()
	if len(register) < 6 { // in case of MIXDuoByte
		register = append([]MIXByte{register[0], 0, 0, 0}, register[1:]...)
	}
	content := machine.ReadCell(inst.A())
	if L == 0 { // if sign is included
		content[0] = register[0]
		L = 1
	}
	copy(content[L:R+1], register[6-(R-L+1):])
	machine.WriteCell(inst.A(), content)
}

var store = InstTemplates{
	"STA": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 24),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.A)
			},
		}
	},
	"ST1": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 25),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I1)
			},
		}
	},
	"ST2": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 26),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I2)
			},
		}
	},
	"ST3": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 27),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I3)
			},
		}
	},
	"ST4": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 28),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I4)
			},
		}
	},
	"ST5": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 29),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I5)
			},
		}
	},
	"ST6": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 30),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.I6)
			},
		}
	},
	"STX": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 31),
			Exec: func(machine *MIXArch, inst *Instruction) {
				ST(machine, inst, machine.Regs.X)
			},
		}
	},
	"STJ": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 32),
			Exec: func(machine *MIXArch, inst *Instruction) {
			},
		}
	},
	"STZ": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 33),
			Exec: func(machine *MIXArch, inst *Instruction) {
			},
		}
	},
}

func aggregateTemplates() InstTemplates {
	templates := make(InstTemplates)
	templates.merge(load)
	return templates
}

var templates = aggregateTemplates()
