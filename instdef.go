package main

type Instruction interface {
	Do(m *MIXArch)
	Fields() MIXBytes
	Duration() int
}

type Add struct {
	fields MIXBytes
}

func newAdd(c MIXByte) *Add {
	return &Add{
		fields: defaultFields(0, 5, c),
	}
}
func (inst *Add) Do(m *MIXArch) {
	data := m.Cell(inst).Slice(FieldSpec(inst))
	if Code(inst) == 2 {
		data = data.Negate()
	}
	sum, overflowed := m.R[A].Raw().Add(data)
	m.OverflowToggle = overflowed
	copy(m.R[A], sum)
}
func (inst *Add) Fields() MIXBytes { return inst.fields }
func (inst *Add) Duration() int    { return 2 }

type Shift struct {
	fields MIXBytes
}

func newShift(R MIXByte) *Shift {
	return &Shift{defaultFields(0, R, 6)}
}
func (inst *Shift) Do(m *MIXArch) {
	rData := make(MIXBytes, 5)
	copy(rData, m.R[A].Data())
	_, R := FieldSpec(inst)
	if 1 < R { // shifts rA + rX (data only, not signs)
		rData = append(rData, m.R[X].Raw().Data()...)
	}
	var (
		shiftAmt = toNum(Address(inst)) % len(rData)
		removed  = make(MIXBytes, shiftAmt)
		vacant   MIXBytes
	)
	if R%2 == 0 { // left shift
		vacant = rData[len(rData)-shiftAmt:]
		copy(removed, rData[:shiftAmt])
		copy(rData, rData[shiftAmt:])
	} else { // right shift
		vacant = rData[:shiftAmt]
		copy(removed, rData[len(rData)-shiftAmt:])
		copy(rData[shiftAmt:], rData[:len(rData)-shiftAmt])
	}

	if 3 < R { // circular
		copy(vacant, removed)
	} else {
		copy(vacant, make(MIXBytes, len(vacant)))
	}
	copy(m.R[A].Data(), rData)
	copy(m.R[X].Data(), rData[5:]) // think nop if rX wasn't included in shift
}
func (inst *Shift) Fields() MIXBytes { return inst.fields }
func (inst *Shift) Duration() int    { return 2 }

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
	data := m.Cell(inst)
	if 15 < Code(inst) {
		data = data.Negate()
	}
	s := data.Slice(FieldSpec(inst))
	dst := m.R[inst.rI]
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

func newST(c MIXByte, rI int) *Store {
	st := &Store{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
	if c == 32 {
		setFieldSpec(st, 0, 2)
	}
	return st
}
func (inst *Store) Do(m *MIXArch) {
	src := m.R[inst.rI]
	switch true {
	case I1 <= inst.rI && inst.rI <= I6:
		src = append(Register{src[0], 0, 0, 0}, src[1:]...)
	case Code(inst) == 33:
		src = Register(NewWord())
	}
	L, R := FieldSpec(inst)
	cell := m.Cell(inst)
	if L == 0 {
		copy(cell.Sign(), src.Raw().Sign())
		L = 1
	}
	copy(cell[L:R+1], src[len(src)-int(R-L+1):])
}
func (inst *Store) Fields() MIXBytes { return inst.fields }
func (inst *Store) Duration() int    { return 2 }

type AddressTransfer struct {
	fields MIXBytes
	rI     int
}

func newAddressTransfer(R, c MIXByte, rI int) *AddressTransfer {
	return &AddressTransfer{
		fields: defaultFields(0, R, c),
		rI:     rI,
	}
}
func (inst *AddressTransfer) Do(m *MIXArch) {
	_, R := FieldSpec(inst)
	address := Address(inst)
	if R%2 == 1 { // DEC, ENN
		address = address.Negate()
	}
	if dst := m.R[inst.rI]; R < 2 { // INC, DEC
		sum, overflowed := dst.Raw().Add(address)
		m.OverflowToggle = overflowed
		copy(dst, sum)
	} else { // ENT, ENN
		copy(dst, address)
	}
}
func (inst *AddressTransfer) Fields() MIXBytes { return inst.fields }
func (inst *AddressTransfer) Duration() int    { return 2 }

type Compare struct {
	fields MIXBytes
	rI     int
}

func newCmp(c MIXByte, rI int) *Compare {
	return &Compare{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
}
func (inst *Compare) Do(m *MIXArch) {
	L, R := FieldSpec(inst)
	rSlice := m.R[inst.rI].Raw().Slice(L, R)
	cellSlice := m.Cell(inst).Slice(L, R)
	rNum, cellNum := toNum(rSlice), toNum(cellSlice)
	m.SetComparisons(rNum < cellNum, rNum == cellNum, rNum > cellNum)
}
func (inst *Compare) Fields() MIXBytes { return inst.fields }
func (inst *Compare) Duration() int    { return 2 }

type Jump struct {
	fields MIXBytes
	rI     int
}

func newJmp(R, c MIXByte, rI int) *Jump {
	return &Jump{
		fields: defaultFields(0, R, c),
		rI:     rI,
	}
}
func (inst *Jump) Do(m *MIXArch) {
	_, R := FieldSpec(inst)
	c, address := Code(inst), Address(inst)

	// comparison flags and values are gathered
	// here to avoid repeating later.
	lt, eq, gt := m.Comparisons()
	var v int64
	if 39 < c {
		v = toNum(m.R[inst.rI].Raw())
	}

	// Jumping seems to consist of writing to
	// rJ and PC.
	setJmp := func() {
		copy(m.R[J], address)
		copy(m.PC, address)
	}

	switch true {
	case c == 39 && R == 0: // JMP
		setJmp()
	case c == 39 && R == 1: // JSJ
		copy(m.PC, address)
	case c == 39 && R == 2: // JOV
		if m.OverflowToggle {
			setJmp()
		}
		m.OverflowToggle = false
	case c == 39 && R == 3: // JNOV
		if !m.OverflowToggle {
			setJmp()
		}
		m.OverflowToggle = false
	case c == 39 && R == 4 && lt: // JL
		setJmp()
	case c == 39 && R == 5 && eq: // JE
		setJmp()
	case c == 39 && R == 6 && gt: // JG
		setJmp()
	case c == 39 && R == 7 && eq && gt: // JGE
		setJmp()
	case c == 39 && R == 8 && lt && gt: // JNE
		setJmp()
	case c == 39 && R == 9 && lt && eq: // JLE
		setJmp()
	case 39 < c && R == 0 && v < 0: // J_N
		setJmp()
	case 39 < c && R == 1 && v == 0: // J_Z
		setJmp()
	case 39 < c && R == 2 && 0 < v: // J_P
		setJmp()
	case 39 < c && R == 3 && -1 < v: // J_NN
		setJmp()
	case 39 < c && R == 4 && v != 0: // J_NZ
		setJmp()
	case 39 < c && R == 5 && v < 1: // J_NP
		setJmp()
	}
}
func (inst *Jump) Fields() MIXBytes { return inst.fields }
func (inst *Jump) Duration() int    { return 2 }

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

	"JMP":  func() Instruction { return newJmp(0, 39, NoR) },
	"JSJ":  func() Instruction { return newJmp(1, 39, NoR) },
	"JOV":  func() Instruction { return newJmp(2, 39, NoR) },
	"JNOV": func() Instruction { return newJmp(3, 39, NoR) },
	"JL":   func() Instruction { return newJmp(4, 39, NoR) },
	"JE":   func() Instruction { return newJmp(5, 39, NoR) },
	"JG":   func() Instruction { return newJmp(6, 39, NoR) },
	"JGE":  func() Instruction { return newJmp(7, 39, NoR) },
	"JNE":  func() Instruction { return newJmp(8, 39, NoR) },
	"JLE":  func() Instruction { return newJmp(9, 39, NoR) },

	"JAN":  func() Instruction { return newJmp(0, 40, A) },
	"JAZ":  func() Instruction { return newJmp(1, 40, A) },
	"JAP":  func() Instruction { return newJmp(2, 40, A) },
	"JANN": func() Instruction { return newJmp(3, 40, A) },
	"JANZ": func() Instruction { return newJmp(4, 40, A) },
	"JANP": func() Instruction { return newJmp(5, 40, A) },

	"J1N":  func() Instruction { return newJmp(0, 41, I1) },
	"J1Z":  func() Instruction { return newJmp(1, 41, I1) },
	"J1P":  func() Instruction { return newJmp(2, 41, I1) },
	"J1NN": func() Instruction { return newJmp(3, 41, I1) },
	"J1NZ": func() Instruction { return newJmp(4, 41, I1) },
	"J1NP": func() Instruction { return newJmp(5, 41, I1) },

	"J2N":  func() Instruction { return newJmp(0, 42, I2) },
	"J2Z":  func() Instruction { return newJmp(1, 42, I2) },
	"J2P":  func() Instruction { return newJmp(2, 42, I2) },
	"J2NN": func() Instruction { return newJmp(3, 42, I2) },
	"J2NZ": func() Instruction { return newJmp(4, 42, I2) },
	"J2NP": func() Instruction { return newJmp(5, 42, I2) },

	"J3N":  func() Instruction { return newJmp(0, 43, I3) },
	"J3Z":  func() Instruction { return newJmp(1, 43, I3) },
	"J3P":  func() Instruction { return newJmp(2, 43, I3) },
	"J3NN": func() Instruction { return newJmp(3, 43, I3) },
	"J3NZ": func() Instruction { return newJmp(4, 43, I3) },
	"J3NP": func() Instruction { return newJmp(5, 43, I3) },

	"J4N":  func() Instruction { return newJmp(0, 44, I4) },
	"J4Z":  func() Instruction { return newJmp(1, 44, I4) },
	"J4P":  func() Instruction { return newJmp(2, 44, I4) },
	"J4NN": func() Instruction { return newJmp(3, 44, I4) },
	"J4NZ": func() Instruction { return newJmp(4, 44, I4) },
	"J4NP": func() Instruction { return newJmp(5, 44, I4) },

	"J5N":  func() Instruction { return newJmp(0, 45, I5) },
	"J5Z":  func() Instruction { return newJmp(1, 45, I5) },
	"J5P":  func() Instruction { return newJmp(2, 45, I5) },
	"J5NN": func() Instruction { return newJmp(3, 45, I5) },
	"J5NZ": func() Instruction { return newJmp(4, 45, I5) },
	"J5NP": func() Instruction { return newJmp(5, 45, I5) },

	"J6N":  func() Instruction { return newJmp(0, 46, I6) },
	"J6Z":  func() Instruction { return newJmp(1, 46, I6) },
	"J6P":  func() Instruction { return newJmp(2, 46, I6) },
	"J6NN": func() Instruction { return newJmp(3, 46, I6) },
	"J6NZ": func() Instruction { return newJmp(4, 46, I6) },
	"J6NP": func() Instruction { return newJmp(5, 46, I6) },

	"JXN":  func() Instruction { return newJmp(0, 47, X) },
	"JXZ":  func() Instruction { return newJmp(1, 47, X) },
	"JXP":  func() Instruction { return newJmp(2, 47, X) },
	"JXNN": func() Instruction { return newJmp(3, 47, X) },
	"JXNZ": func() Instruction { return newJmp(4, 47, X) },
	"JXNP": func() Instruction { return newJmp(5, 47, X) },

	"INCA": func() Instruction { return newAddressTransfer(0, 48, A) },
	"INC1": func() Instruction { return newAddressTransfer(0, 49, I1) },
	"INC2": func() Instruction { return newAddressTransfer(0, 50, I2) },
	"INC3": func() Instruction { return newAddressTransfer(0, 51, I3) },
	"INC4": func() Instruction { return newAddressTransfer(0, 52, I4) },
	"INC5": func() Instruction { return newAddressTransfer(0, 53, I5) },
	"INC6": func() Instruction { return newAddressTransfer(0, 54, I6) },
	"INCX": func() Instruction { return newAddressTransfer(0, 55, X) },

	"DECA": func() Instruction { return newAddressTransfer(1, 48, A) },
	"DEC1": func() Instruction { return newAddressTransfer(1, 49, I1) },
	"DEC2": func() Instruction { return newAddressTransfer(1, 50, I2) },
	"DEC3": func() Instruction { return newAddressTransfer(1, 51, I3) },
	"DEC4": func() Instruction { return newAddressTransfer(1, 52, I4) },
	"DEC5": func() Instruction { return newAddressTransfer(1, 53, I5) },
	"DEC6": func() Instruction { return newAddressTransfer(1, 54, I6) },
	"DECX": func() Instruction { return newAddressTransfer(1, 55, X) },

	"ENTA": func() Instruction { return newAddressTransfer(2, 48, A) },
	"ENT1": func() Instruction { return newAddressTransfer(2, 49, I1) },
	"ENT2": func() Instruction { return newAddressTransfer(2, 50, I2) },
	"ENT3": func() Instruction { return newAddressTransfer(2, 51, I3) },
	"ENT4": func() Instruction { return newAddressTransfer(2, 52, I4) },
	"ENT5": func() Instruction { return newAddressTransfer(2, 53, I5) },
	"ENT6": func() Instruction { return newAddressTransfer(2, 54, I6) },
	"ENTX": func() Instruction { return newAddressTransfer(2, 55, X) },

	"ENNA": func() Instruction { return newAddressTransfer(2, 48, A) },
	"ENN1": func() Instruction { return newAddressTransfer(2, 49, I1) },
	"ENN2": func() Instruction { return newAddressTransfer(2, 50, I2) },
	"ENN3": func() Instruction { return newAddressTransfer(2, 51, I3) },
	"ENN4": func() Instruction { return newAddressTransfer(2, 52, I4) },
	"ENN5": func() Instruction { return newAddressTransfer(2, 53, I5) },
	"ENN6": func() Instruction { return newAddressTransfer(2, 54, I6) },
	"ENNX": func() Instruction { return newAddressTransfer(2, 55, X) },
}
