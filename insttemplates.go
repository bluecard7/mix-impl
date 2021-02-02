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
		templates["LD"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						ld(inst, m.R[regI], m.Cell(inst.A()))
					},
				}
			}
		}(8+i, entry.I)

		templates["LD"+entry.Name+"N"] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						ld(inst, m.R[regI], m.Cell(inst.A()).Negate())
					},
				}
			}
		}(16+i, entry.I)
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
		templates["ST"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				op := MIXByte(c)
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
		}(24+i, entry.I)
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
		templates["INC"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 0, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						sum, overflowed := m.R[regI].Raw().Add(inst.A())
						m.OverflowToggle = overflowed
						copy(m.R[regI], sum)
					},
				}
			}
		}(48+i, entry.I)
		// DEC, F = 1
		templates["DEC"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 1, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						sum, overflowed := m.R[regI].Raw().Add(inst.A().Negate())
						m.OverflowToggle = overflowed
						copy(m.R[regI], sum)
					},
				}
			}
		}(48+i, entry.I)
		// ENT, F = 2
		templates["ENT"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 2, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						copy(m.R[regI], inst.A())
					},
				}
			}
		}(48+i, entry.I)
		//ENN, F = 3
		templates["ENN"+entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 3, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						copy(m.R[regI], inst.A().Negate())
					},
				}
			}
		}(48+i, entry.I)
	}
	return templates
}

func comparisons() InstTemplates {
	templates := make(InstTemplates)
	for i, entry := range regEntries {
		templates[entry.Name] = func(c, regI int) func() *Instruction {
			return func() *Instruction {
				return &Instruction{
					Code: baseInstCode(0, 5, MIXByte(c)),
					Exec: func(m *MIXArch, inst *Instruction) {
						rS := m.R[regI].Raw().Slice(inst.F())
						cellS := m.Cell(inst.A()).Slice(inst.F())
						rNum, cellNum := toNum(rS), toNum(cellS)
						m.SetComparisons(rNum < cellNum, rNum == cellNum, rNum > cellNum)
					},
				}
			}
		}(56+i, entry.I)
	}
	return templates
}

func jumps() InstTemplates {
	templates := make(InstTemplates)
	templates["JMP"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 0, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				copy(m.R[J], inst.A())
				copy(m.PC, inst.A())
			},
		}
	}
	templates["JSJ"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 1, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				copy(m.PC, inst.A())
			},
		}
	}
	templates["JOV"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 2, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if m.OverflowToggle {
					m.OverflowToggle = false
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JNOV"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 3, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if !m.OverflowToggle {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
				m.OverflowToggle = false
			},
		}
	}
	// JL, JE, JG, JGE, JNE, JLE
	templates["JL"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 4, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if lt, _, _ := m.Comparisons(); lt {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JE"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if _, eq, _ := m.Comparisons(); eq {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JG"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 6, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if _, _, gt := m.Comparisons(); gt {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JGE"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 7, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if _, eq, gt := m.Comparisons(); eq || gt {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JNE"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(1, 0, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if lt, _, gt := m.Comparisons(); lt || gt {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	templates["JLE"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(1, 1, 39),
			Exec: func(m *MIXArch, inst *Instruction) {
				if lt, eq, _ := m.Comparisons(); lt || eq {
					copy(m.R[J], inst.A())
					copy(m.PC, inst.A())
				}
			},
		}
	}
	for i, entry := range regEntries {
		for j, suffix := range []string{"N", "Z", "P", "NN", "NZ", "NP"} {
			templates["J"+entry.Name+suffix] = func(c, f, regI int) func() *Instruction {
				return func() *Instruction {
					return &Instruction{
						Code: baseInstCode(0, MIXByte(f), MIXByte(c)),
						Exec: func(m *MIXArch, inst *Instruction) {
							if n := toNum(m.R[regI].Raw()); (f == 0 && n < 0) || (f == 1 && n == 0) ||
								(f == 2 && 0 < n) || (f == 3 && 0 <= n) ||
								(f == 4 && n != 0) || (f == 5 && n <= 0) {
								copy(m.R[J], inst.A())
								copy(m.PC, inst.A())
							}
						},
					}
				}
			}(40+i, j, entry.I)
		}
	}
	return templates
}

func shifts() InstTemplates {
	return InstTemplates{}
}

func aggregateTemplates() InstTemplates {
	templates := make(InstTemplates)
	templates["NOP"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 0, 0),
			Exec: func(m *MIXArch, inst *Instruction) {},
		}
	}
	templates["HLT"] = func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 2, 5),
			Exec: func(m *MIXArch, inst *Instruction) {
				// stop machine, involve program counter?
			},
		}
	}
	templates.merge(loads())
	templates.merge(stores())
	templates.merge(arithmetic())
	return templates
}

var templates = aggregateTemplates()
