package main

type InstTemplates map[string]func() *Instruction

func (dst InstTemplates) merge(src InstTemplates) {
	for k, v := range src {
		dst[k] = v
	}
}

func LD(inst *Instruction, dst Register, content []MIXByte) {
	L, R := inst.F()
	dst[0] = POS_SIGN
	if L == 0 {
		dst[0] = content[0]
		L = 1
	}
	partial, amtToCpy, actualSize := content[L:R+1], R-L+1, R-L+1
	if len(dst) < WORD_SIZE && 2 < amtToCpy { // index registers
		amtToCpy = 2
	}
	copy(dst[len(dst)-int(amtToCpy):], partial[actualSize-amtToCpy:])
}

// load insts, #8 - 23
func loads() InstTemplates {
	templates := make(InstTemplates)
	entries := []struct {
		Name string
		R    int
	}{
		{"LDA", A},
		{"LD1", I1},
		{"LD2", I2},
		{"LD3", I3},
		{"LD4", I4},
		{"LD5", I5},
		{"LD6", I6},
		{"LDX", X},
		{"LDAN", A}, // load negative versions
		{"LD1N", I1},
		{"LD2N", I2},
		{"LD3N", I3},
		{"LD4N", I4},
		{"LD5N", I5},
		{"LD6N", I6},
		{"LDXN", X},
	}
	for i, entry := range entries {
		templates[entry.Name] = func(opOffset, regI int) func() *Instruction {
			return func() *Instruction {
				op := MIXByte(8 + opOffset)
				return &Instruction{
					Code: baseInstCode(0, 5, op),
					Exec: func(machine *MIXArch, inst *Instruction) {
						content := machine.ReadCell(inst.A())
						if 15 < op {
							content.negate()
						}
						LD(inst, machine.R[regI], content)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

func ST(machine *MIXArch, inst *Instruction, src Register) {
	L, R := inst.F()
	content := machine.ReadCell(inst.A())
	if L == 0 { // if sign is included
		content[0] = src[0]
		L = 1
	}
	copy(content[L:R+1], src[len(src)-int(R-L+1):])
	machine.WriteCell(inst.A(), content)
}

// store insts, #24-33
func stores() InstTemplates {
	entries := []struct {
		Name string
		R    int
	}{
		{"STA", A},
		{"ST1", I1},
		{"ST2", I2},
		{"ST3", I3},
		{"ST4", I4},
		{"ST5", I5},
		{"ST6", I6},
		{"STX", X},
		{"STJ", J},
		{"STZ", A},
	}
	templates := make(InstTemplates)
	for i, entry := range entries {
		templates[entry.Name] = func(opOffset, regI int) func() *Instruction {
			return func() *Instruction {
				op := MIXByte(24 + opOffset)
				var L, R MIXByte = 0, 5
				if op == 32 { // STJ
					R = 2
				}
				return &Instruction{
					Code: baseInstCode(L, R, op),
					Exec: func(machine *MIXArch, inst *Instruction) {
						r := machine.R[regI]
						if I1 <= regI || regI <= I6 { // using index register
							r = append([]MIXByte{r[0], 0, 0, 0}, r[1:]...)
						}
						if op == 33 { // STZ
							r = Register(NewWord())
						}
						ST(machine, inst, r)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

func aggregateTemplates() InstTemplates {
	templates := make(InstTemplates)
	templates.merge(loads())
	templates.merge(stores())
	return templates
}

var templates = aggregateTemplates()
