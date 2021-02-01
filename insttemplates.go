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

// Bundles index of register with suffix used in instructions affecting the register
type RegEntry struct {
	Name string
	I    int
}

var regEntries = []RegEntry{
	{"A", A}, {"1", I1}, {"2", I2}, {"3", I3}, {"4", I4}, {"5", I5}, {"6", I6}, {"X", X},
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
	for i, entry := range regEntries {
		templates["LD"+entry.Name] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(8+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						ld(inst, m.R[regI], m.Cell(inst.A()))
					},
				}
			}
		}(i, entry.I)

		templates["LD"+entry.Name+"N"] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(16+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						ld(inst, m.R[regI], m.Cell(inst.A()).Negate())
					},
				}
			}
		}(i, entry.I)
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
	stEntries := append(regEntries, RegEntry{"J", J}, RegEntry{"Z", A})
	templates := make(InstTemplates)
	for i, entry := range stEntries {
		templates["ST"+entry.Name] = func(opOffset, regI int) func() *Instruction {
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
		}(i, entry.I)
	}
	return templates
}

// add, sub, mul, div, #1-4
func arithmetic() InstTemplates {
	return InstTemplates{
		"ADD": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 1),
				Exec: func(m *MIXArch, inst *Instruction) {
					cell := m.Cell(inst.A()).Slice(inst.F())
					sum, overflowed := m.R[A].Raw().Add(cell)
					m.OverflowToggle = overflowed
					copy(m.R[A], sum)
				},
			}
		},
		"SUB": func() *Instruction {
			return &Instruction{
				Code: baseInstCode(0, 5, 2),
				Exec: func(m *MIXArch, inst *Instruction) {
					cell := m.Cell(inst.A()).Slice(inst.F())
					sum, overflowed := m.R[A].Raw().Add(cell.Negate())
					m.OverflowToggle = overflowed
					copy(m.R[A], sum)
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
	templates := make(InstTemplates)
	for i, entry := range regEntries {
		// INC, F = 0
		templates["INC"+entry.Name] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 0, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						sum, overflowed := m.R[regI].Raw().Add(inst.A())
						m.OverflowToggle = overflowed
						copy(m.R[regI], sum)
					},
				}
			}
		}(i, entry.I)
		// DEC, F = 1
		templates["DEC"+entry.Name] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 1, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						sum, overflowed := m.R[regI].Raw().Add(inst.A().Negate())
						m.OverflowToggle = overflowed
						copy(m.R[regI], sum)
					},
				}
			}
		}(i, entry.I)
		// ENT, F = 2
		templates["ENT"+entry.Name] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 2, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						copy(m.R[regI], inst.A())
					},
				}
			}
		}(i, entry.I)
		//ENN, F = 3
		templates["ENN"+entry.Name] = func(cOffset, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 3, MIXByte(48+cOffset)),
					Exec: func(m *MIXArch, inst *Instruction) {
						copy(m.R[regI], inst.A().Negate())
					},
				}
			}
		}(i, entry.I)
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
