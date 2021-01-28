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
	partial := content[L : R+1]
	copy(dst[len(dst)-len(partial):], partial)
}

// load insts, #8 - 15
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
	}
	for i, entry := range entries {
		templates[entry.Name] = func(codeOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(8+codeOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
						content := machine.ReadCell(inst.A())
						LD(inst, machine.R[regI], content)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

// load negative insts, #16 - 23
func loadNs() InstTemplates {
	templates := make(InstTemplates)
	entries := []struct {
		Name string
		R    int
	}{
		{"LDAN", A},
		{"LD1N", I1},
		{"LD2N", I2},
		{"LD3N", I3},
		{"LD4N", I4},
		{"LD5N", I5},
		{"LD6N", I6},
		{"LDXN", X},
	}
	for i, entry := range entries {
		templates[entry.Name] = func(codeOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(16+codeOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
						content := machine.ReadCell(inst.A())
						switch content[0] {
						case POS_SIGN:
							content[0] = NEG_SIGN
						case NEG_SIGN:
							content[0] = POS_SIGN
						}
						LD(inst, machine.R[regI], content)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

func ST(machine *MIXArch, inst *Instruction, r Register) {
	L, R := inst.F()
	if len(r) < 6 { // in case of index or jump register
		r = append([]MIXByte{r[0], 0, 0, 0}, r[1:]...)
	}
	content := machine.ReadCell(inst.A())
	if L == 0 { // if sign is included
		content[0] = r[0]
		L = 1
	}
	copy(content[L:R+1], r[6-(R-L+1):])
	machine.WriteCell(inst.A(), content)
}

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
		templates[entry.Name] = func(codeOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(24+codeOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
						r := machine.R[regI]
						if entry.Name == "STZ" {
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
	templates.merge(loadNs())
	templates.merge(stores())
	return templates
}

var templates = aggregateTemplates()
