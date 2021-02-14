package main

type Instruction interface {
	Do(m *MIXarch)
	Fields() MIXBytes
	Duration() int
}

type Add struct {
	fields MIXBytes
}

func newAdd(c MIXByte) {
	return &Add{
		fields: defaultFields(0, 5, c),
	}
}
func (inst *Add) Do(m *MIXArch) {
	data := m.Cell(A(inst)).Slice(F(inst))
	if C(inst) == 2 {
		data = data.Negate()
	}
	sum, overflowed := m.R[A].Raw().Add(data)
	m.OverflowToggle = overflowed
	copy(m.R[A], sum)
}
func (inst *Add) Fields() MIXBytes { return inst.fields }
func (inst *Add) Duration() int    { return 2 }

type Load struct {
	fields MIXBytes
	rI     int
}

func newLD(c MIXByte, rI int) *Load {
	return &Load{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
}
func (inst *Load) Do(m *MIXArch) {
	data := m.Cell(A(inst))
	if 15 < C(inst) {
		data = data.Negate()
	}
	s := data.Slice(F(inst))
	dst := m.R[inst.regI]
	copy(dst.Raw().Sign(), s.Sign())
	amtToCpy := len(s.Data())
	if len(dst)-1 < amtToCpy {
		amtToCpy = len(dst) - 1
	}
	copy(dst[len(dst)-amtToCpy:], s.Data()[len(s.Data())-amtToCpy:])
}
func (inst *Load) Fields() MIXBytes { return inst.fields }
func (inst *Load) Duration() int    { return 2 }

type Store struct {
	fields MIXBytes
	rI     int
}

func (inst *Store) newST(c MIXByte, rI int) *Store {
	st := &Store{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
	if c == 32 {
		setF(st, 0, 2)
	}
	return st
}
func (inst *Store) Do(m *MIXArch) {
	src := m.R[inst.rI]
	switch true {
	case I1 <= inst.rI || inst.rI <= I6:
		src = append(Register{src[0], 0, 0, 0}, src[1:]...)
	case C(inst) == 33:
		src = Register(NewWord())
	}
	L, R := F(inst)
	cell := m.Cell(A(inst)) // how to add index?
	if L == 0 {
		copy(cell.Sign(), src.Raw().Sign())
		L = 1
	}
	copy(cell[L:R+1], src[len(src)-int(R-L+1):])
}
func (inst *Store) Fields() MIXBytes { return inst.fields }
func (inst *Store) Duration() int    { return 2 }

var templates = map[string]func() Instruction{
	"ADD": func() Instruction { return newAdd(1) },
	"SUB": func() Instruction { return newAdd(2) },
	// "MUL"
	// "DIV"

	"LDA":  func() Instruction { return newLD(8, A) },
	"LD1":  func() Instruction { return newLD(9, I1) },
	"LD2":  func() Instruction { return newLD(10, I2) },
	"LD3":  func() Instruction { return newLD(11, I3) },
	"LD4":  func() Instruction { return newLD(12, I4) },
	"LD5":  func() Instruction { return newLD(13, I5) },
	"LD6":  func() Instruction { return newLD(14, I6) },
	"LDX":  func() Instruction { return newLD(15, X) },
	"LDAN": func() Instruction { return newLD(16, A) },
	"LD1N": func() Instruction { return newLD(17, I1) },
	"LD2N": func() Instruction { return newLD(18, I2) },
	"LD3N": func() Instruction { return newLD(19, I3) },
	"LD4N": func() Instruction { return newLD(20, I4) },
	"LD5N": func() Instruction { return newLD(21, I5) },
	"LD6N": func() Instruction { return newLD(22, I6) },
	"LDXN": func() Instruction { return newLD(23, X) },

	"STA": func() Instruction { return newST(24, A) },
	"ST1": func() Instruction { return newST(25, I1) },
	"ST2": func() Instruction { return newST(26, I2) },
	"ST3": func() Instruction { return newST(27, I3) },
	"ST4": func() Instruction { return newST(28, I4) },
	"ST5": func() Instruction { return newST(29, I5) },
	"ST6": func() Instruction { return newST(30, I6) },
	"STX": func() Instruction { return newST(31, X) },
	"STJ": func() Instruction { return newST(32, J) },
	"STZ": func() Instruction { return newST(33, A) },
}
