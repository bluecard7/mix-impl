package main

import (
	"math"
)

type InstTemplates map[string]func() *Instruction

func (dst InstTemplates) merge(src InstTemplates) {
	for k, v := range src {
		dst[k] = v
	}
}

func ld(inst *Instruction, dst Register, data MIXBytes) {
	s := data.Slice(inst.F()) // guaranteed sign
	copy(dst.Raw().Sign(), s.Sign())
	amtToCpy := int(math.Min(
		float64(len(s.Data())),
		float64(len(dst)-1),
	))
	copy(dst[len(dst)-amtToCpy:], s.Data()[len(s.Data())-amtToCpy:])
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
					Exec: func(m *MIXArch, inst *Instruction) {
						cell := m.Cell(inst.A())
						if 15 < op {
							cell.Negate()
						}
						ld(inst, m.R[regI], cell)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

func st(m *MIXArch, inst *Instruction, src Register) {
	L, R := inst.F()
	cell := m.Cell(inst.A())
	if L == 0 {
		copy(cell.Sign(), src.Raw().Sign())
		L++
	}
	copy(cell[L:R+1], src[len(src)-int(R-L+1):])
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
					Exec: func(m *MIXArch, inst *Instruction) {
						r := m.R[regI]
						if I1 <= regI || regI <= I6 { // using index register
							r = append(Register{r[0], 0, 0, 0}, r[1:]...)
						}
						if op == 33 { // STZ
							r = Register(NewWord())
						}
						st(m, inst, r)
					},
				}
			}
		}(i, entry.R)
	}
	return templates
}

// add, sub, mul, div, #1-4
func arithmetic() InstTemplates {
	mixSum := func(m *MIXArch, n1, n2 MIXBytes) MIXBytes {
		sum := toNum(n1) + toNum(n2)
		if sum > 2<<31-1 {
			m.OverflowToggle = true
		}
		return toMIXBytes(sum, 5)
	}

	return InstTemplates{
		"ADD": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 1),
				Exec: func(m *MIXArch, inst *Instruction) {
					cell := m.Cell(inst.A()).Slice(inst.F())
					copy(m.R[A], mixSum(m, m.R[A].Raw(), cell))
				},
			}
		},
		"SUB": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 2),
				Exec: func(m *MIXArch, inst *Instruction) {
					cell := m.Cell(inst.A()).Slice(inst.F())
					copy(m.R[A], mixSum(m, m.R[A].Raw(), cell.Negate()))
				},
			}
		},
		"MUL": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 3),
				Exec: func(m *MIXArch, inst *Instruction) {
					cell := m.Cell(inst.A()).Slice(inst.F())
					product := toNum(m.R[A].Raw()) * toNum(cell)
					productBytes := toMIXBytes(product, 10)
					copy(m.R[X], productBytes[:6])
					copy(m.R[A].Data(), productBytes[6:])
					copy(m.R[A], productBytes.Data())
				},
			}
		},
		"DIV": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 4),
				Exec: func(m *MIXArch, inst *Instruction) {
					var quotient, remainder int64
					cell := m.Cell(inst.A()).Slice(inst.F())
					den := toNum(cell)
					if den != 0 {
						numBytes := append(m.R[A].Sign(), append(m.R[X].Data(), m.R[A].Data()...)...)
						num := toNum(numBytes)
						quotient, remainder = num/den, num%den
					}
					if den == 0 || quotient > 2<<31-1 {
						m.OverflowToggle = true
						return
					}
					copy(m.R[X], toMIXBytes(remainder, 5))
					copy(m.R[X], m.R[A].Sign())
					copy(m.R[A], toMIXBytes(quotient, 5))
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
					Exec: func(m *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		// DEC, F = 1
		templates["DEC"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 1, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		// ENT, F = 2
		templates["ENT"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 2, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
					},
				}
			}
		}(i, entry.RegI)
		//ENN, F = 3
		templates["ENN"+entry.Suffix] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 3, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
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
