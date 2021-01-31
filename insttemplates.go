package main

type InstTemplates map[string]func() *Instruction

func (dst InstTemplates) merge(src InstTemplates) {
	for k, v := range src {
		dst[k] = v
	}
}

func ld(inst *Instruction, dst Register, content MIXBytes) {
	L, R := inst.F()
	dst[0] = POS_SIGN
	if L == 0 {
		dst[0] = content[0]
		L = 1
	}
	partial, amtToCpy, actualSize := content[L:R+1], int(R-L+1), int(R-L+1)
	if len(dst) < WORD_SIZE && len(dst)-1 < amtToCpy { // index registers
		amtToCpy = len(dst) - 1
	}
	copy(dst[len(dst)-amtToCpy:], partial[actualSize-amtToCpy:])
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
							content.Negate()
						}
						ld(inst, machine.R[regI], content)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

func st(machine *MIXArch, inst *Instruction, src Register) {
	L, R := inst.F()
	cell := machine.ReadCell(inst.A()) // is ReadCell really right? it returns the cell
	if L == 0 {                        // if sign is included
		cell[0] = src[0]
		L = 1
	}
	copy(cell[L:R+1], src[len(src)-int(R-L+1):])
	//machine.WriteCell(inst.A(), content)
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
							r = append(Register{r[0], 0, 0, 0}, r[1:]...)
						}
						if op == 33 { // STZ
							r = Register(NewWord())
						}
						st(machine, inst, r)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

// add, sub, mul, div, #1-4
func arithmetic() InstTemplates {
	// Returns MIXBytes of cell specified by address and fields.
	// Makes sure returned MIXBytes has a sign.
	numCell := func(machine *MIXArch, inst *Instruction) MIXBytes {
		L, R := inst.F()
		v := machine.ReadCell(inst.A())[L : R+1]
		if 0 < L {
			v = append(MIXBytes{POS_SIGN}, v...)
		}
		return v
	}

	mixSum := func(machine *MIXArch, n1, n2 MIXBytes) MIXBytes {
		sum := toNum(n1) + toNum(n2)
		if sum > 2<<31-1 {
			machine.OverflowToggle = true
		}
		return toMIXBytes(sum, 5)
	}

	return InstTemplates{
		"ADD": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 1),
				Exec: func(machine *MIXArch, inst *Instruction) {
					copy(machine.R[A], mixSum(machine, machine.R[A].Raw(), numCell(machine, inst)))
				},
			}
		},
		"SUB": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 2),
				Exec: func(machine *MIXArch, inst *Instruction) {
					copy(machine.R[A], mixSum(machine, machine.R[A].Raw(), numCell(machine, inst).Negate()))
				},
			}
		},
		"MUL": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 3),
				Exec: func(machine *MIXArch, inst *Instruction) {
					product := toNum(machine.R[A].Raw()) * toNum(numCell(machine, inst))
					productBytes := toMIXBytes(product, 10)
					copy(machine.R[X], productBytes[:6])
					copy(machine.R[A][1:], productBytes[6:])
					copy(machine.R[A], productBytes[:1])
				},
			}
		},
		"DIV": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 4),
				Exec: func(machine *MIXArch, inst *Instruction) {
					var quotient, remainder int64
					den := toNum(numCell(machine, inst))
					if den != 0 {
						numBytes := append(
							machine.R[A][:1],
							append(machine.R[X][1:], machine.R[A][1:]...)...,
						)
						num := toNum(numBytes.Raw())
						quotient, remainder = num/den, num%den
					}
					if den == 0 || quotient > 2<<31-1 {
						machine.OverflowToggle = true
						return
					}
					copy(machine.R[X], toMIXBytes(remainder, 5))
					copy(machine.R[X], machine.R[A][:1])
					copy(machine.R[A], toMIXBytes(quotient, 5))
				},
			}
		},
	}
}

// #48-55
func addressTransfers() InstTemplates {
	entries := []struct {
		Suffix string
		RegI   int
	}{{"A", A}, {"1", I1}, {"2", I2}, {"3", I3}, {"4", I4}, {"5", I5}, {"6", I6}, {"X", X}}
	templates := make(InstTemplates)
	for i, entry := range entries {
		// INC, F = 0
		templates["INC"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 0, MIXByte(48+cOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		// DEC, F = 1
		templates["DEC"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 1, MIXByte(48+cOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		// ENT, F = 2
		templates["ENT"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 2, MIXByte(48+cOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		//ENN, F = 3
		templates["ENN"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 3, MIXByte(48+cOffset)),
					Exec: func(machine *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
	}
	return templates
}

func aggregateTemplates() InstTemplates {
	templates := make(InstTemplates)
	templates.merge(loads())
	templates.merge(stores())
	templates.merge(arithmetic())
	return templates
}

var templates = aggregateTemplates()
