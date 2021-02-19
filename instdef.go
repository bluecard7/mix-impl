package main

type Instruction interface {
	Effect(m *MIXArch) *Snapshot
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
func (inst *Add) Effect(m *MIXArch) *Snapshot {
	data := m.Cell(inst).Slice(FieldSpec(inst))
	if Code(inst) == 2 {
		data = data.Negate()
	}
	sum, overflowed := m.R[A].Raw().Add(data)
	m.OverflowToggle = overflowed
	copy(m.R[A], sum)

	snapshot := new(Snapshot)
	snapshot.includesR(A, m.R[A])
	return snapshot
}
func (inst *Add) Fields() MIXBytes { return inst.fields }
func (inst *Add) Duration() int    { return 2 }

type Convert struct {
	fields MIXBytes
}

func newConv(R MIXByte) *Convert {
	return &Convert{defaultFields(0, R, 5)}
}

// TODO:: conversions NUM and CHAR

type Shift struct {
	fields MIXBytes
}

func newShift(R MIXByte) *Shift {
	return &Shift{defaultFields(0, R, 6)}
}
func (inst *Shift) Effect(m *MIXArch) *Snapshot {
	snapshot := new(Snapshot)
	defer func() { snapshot.includesR(A, m.R[A]) }()
	rData := make(MIXBytes, 5)
	copy(rData, m.R[A].Data())
	_, R := FieldSpec(inst)
	if 1 < R { // shifts rA + rX (data only, not signs)
		rData = append(rData, m.R[X].Raw().Data()...)
		defer func() { snapshot.includesR(X, m.R[X]) }()
	}
	var (
		size     = len(rData)
		shiftAmt = toNum(Address(inst)) % size
		removed  = make(MIXBytes, shiftAmt)
		vacant   MIXBytes
	)
	if R%2 == 0 { // left shift
		vacant = rData[size-shiftAmt:]
		copy(removed, rData[:shiftAmt])
		copy(rData, rData[shiftAmt:])
	} else { // right shift
		vacant = rData[:shiftAmt]
		copy(removed, rData[size-shiftAmt:])
		copy(rData[shiftAmt:], rData[:size-shiftAmt])
	}

	if 3 < R { // circular
		copy(vacant, removed)
	} else {
		copy(vacant, make(MIXBytes, len(vacant)))
	}
	copy(m.R[A].Data(), rData)
	copy(m.R[X].Data(), rData[5:]) // think nop if rX wasn't included in shift
	return snapshot
}
func (inst *Shift) Fields() MIXBytes { return inst.fields }
func (inst *Shift) Duration() int    { return 2 }

type Move struct {
	fields MIXBytes
}

func newMove(F MIXByte) *Move {
	// weird to put F here, but once extracted, it will be "fine"
	return &Move{fields: defaultFields(0, F, 7)}
}

func (inst *Move) Effect(m *MIXArch) *Snapshot {
	snapshot := new(Snapshot)
	L, R := FieldSpec(inst)
	mvAmt, srcI, dstI := int(8*L+R), toNum(Address(inst)), toNum(m.R[I1].Raw())
	if srcI != dstI { // nop otherwise, copies cell to itself
		for i := 0; i < mvAmt; i++ {
			copy(m.Mem[dstI+i], m.Mem[srcI+i])
			defer func() { snapshot.includesCell(dstI+i, m.Mem[dstI+i]) }()
		}
		copy(m.R[I1], toMIXBytes(dstI+mvAmt, 2))
		defer func() { snapshot.includesR(I1, m.R[I1]) }()
	}
	return snapshot
}
func (inst *Move) Fields() MIXBytes { return inst.fields }
func (inst *Move) Duration() int    { return 2 }

type Load struct {
	fields MIXBytes
	rI     MIXByte
}

func newLD(c, rI MIXByte) *Load {
	return &Load{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
}
func (inst *Load) Effect(m *MIXArch) *Snapshot {
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

	snapshot := new(Snapshot)
	snapshot.includesR(int(inst.rI), dst)
	return snapshot
}
func (inst *Load) Fields() MIXBytes { return inst.fields }
func (inst *Load) Duration() int    { return 2 }

type Store struct {
	fields MIXBytes
	rI     MIXByte
}

func newST(c, rI MIXByte) *Store {
	st := &Store{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
	if c == 32 {
		setFieldSpec(st, 0, 2)
	}
	return st
}
func (inst *Store) Effect(m *MIXArch) *Snapshot {
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
	snapshot := new(Snapshot)
	snapshot.includesCell(int(toNum(Address(inst))), cell)
	return snapshot
}
func (inst *Store) Fields() MIXBytes { return inst.fields }
func (inst *Store) Duration() int    { return 2 }

type IO struct {
	fields MIXBytes
}

func newIO(R, c MIXByte) *IO {
	return &IO{defaultFields(0, R, c)}
}
func (inst *IO) Effect(m *MIXArch) *Snapshot {
	/*
		// Involves rX somehow with posiitoning device
		L, R := FieldSpec(inst)
		// need to attach device to machine
		device := m.Devices[8*L+R]
		start := toNum(Address(inst))
		end := start + len(device) + 1
		switch Code(inst) {
		case 35: // IOC
		m.Devices.Control(device) or device.Reset if behavior is pretty uniform
		case 36: // IN
		// Not really what I want, copies over MIXBytes
		copy(m.Mem[start:end], device)
		// make some device.Write(m.Mem, start)
		case 37: // OUT
		// device.Read(m.Mem, start)
		case 34: // JBUS
		case 38: // JRED
		}
	*/
	return new(Snapshot)
}
func (inst *IO) Fields() MIXBytes { return inst.fields }
func (inst *IO) Duration() int    { return 2 }

type Jump struct {
	fields MIXBytes
	rI     MIXByte
}

func newJmp(R, c, rI MIXByte) *Jump {
	return &Jump{
		fields: defaultFields(0, R, c),
		rI:     rI,
	}
}
func (inst *Jump) Effect(m *MIXArch) *Snapshot {
	_, R := FieldSpec(inst)
	c, address := Code(inst), Address(inst)

	// comparison flags and values are gathered
	// here to avoid repeating later.
	lt, eq, gt := m.Comparisons()
	var v int
	if 39 < c {
		v = toNum(m.R[inst.rI].Raw())
	}

	// Jumping seems to consist of writing to
	// rJ and PC.
	_setJmp := func() {
		copy(m.R[J], address)
		m.PC = toNum(address)
	}

	snapshot := new(Snapshot)
	isSet := true

	switch true {
	case c == 39 && R == 0: // JMP
		_setJmp()
	case c == 39 && R == 1: // JSJ
		m.PC = toNum(address)
		isSet = false
	case c == 39 && R == 2: // JOV
		if m.OverflowToggle {
			_setJmp()
		} else {
			isSet = false
		}
		m.OverflowToggle = false
	case c == 39 && R == 3: // JNOV
		if !m.OverflowToggle {
			_setJmp()
		} else {
			isSet = false
		}
		m.OverflowToggle = false
	case c == 39 && R == 4 && lt: // JL
		_setJmp()
	case c == 39 && R == 5 && eq: // JE
		_setJmp()
	case c == 39 && R == 6 && gt: // JG
		_setJmp()
	case c == 39 && R == 7 && eq && gt: // JGE
		_setJmp()
	case c == 39 && R == 8 && lt && gt: // JNE
		_setJmp()
	case c == 39 && R == 9 && lt && eq: // JLE
		_setJmp()
	case 39 < c && R == 0 && v < 0: // J_N
		_setJmp()
	case 39 < c && R == 1 && v == 0: // J_Z
		_setJmp()
	case 39 < c && R == 2 && 0 < v: // J_P
		_setJmp()
	case 39 < c && R == 3 && -1 < v: // J_NN
		_setJmp()
	case 39 < c && R == 4 && v != 0: // J_NZ
		_setJmp()
	case 39 < c && R == 5 && v < 1: // J_NP
		_setJmp()
	}
	if isSet {
		snapshot.includesR(J, m.R[J])
	}
	return snapshot
}
func (inst *Jump) Fields() MIXBytes { return inst.fields }
func (inst *Jump) Duration() int    { return 2 }

type AddressTransfer struct {
	fields MIXBytes
	rI     MIXByte
}

func newAddressTransfer(R, c, rI MIXByte) *AddressTransfer {
	return &AddressTransfer{
		fields: defaultFields(0, R, c),
		rI:     rI,
	}
}
func (inst *AddressTransfer) Effect(m *MIXArch) *Snapshot {
	_, R := FieldSpec(inst)
	address := Address(inst)
	if R%2 == 1 { // DEC, ENN
		address = address.Negate()
	}
	dst := m.R[inst.rI]
	if R < 2 { // INC, DEC
		sum, overflowed := dst.Raw().Add(address)
		m.OverflowToggle = overflowed
		copy(dst, sum)
	} else { // ENT, ENN
		copy(dst, address)
	}
	snapshot := new(Snapshot)
	snapshot.includesR(int(inst.rI), dst)
	return snapshot
}
func (inst *AddressTransfer) Fields() MIXBytes { return inst.fields }
func (inst *AddressTransfer) Duration() int    { return 2 }

type Compare struct {
	fields MIXBytes
	rI     MIXByte
}

func newCmp(c, rI MIXByte) *Compare {
	return &Compare{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
}
func (inst *Compare) Effect(m *MIXArch) *Snapshot {
	L, R := FieldSpec(inst)
	rSlice := m.R[inst.rI].Raw().Slice(L, R)
	cellSlice := m.Cell(inst).Slice(L, R)
	rNum, cellNum := toNum(rSlice), toNum(cellSlice)
	m.SetComparisons(rNum < cellNum, rNum == cellNum, rNum > cellNum)
	return new(Snapshot) // include comparators?
}
func (inst *Compare) Fields() MIXBytes { return inst.fields }
func (inst *Compare) Duration() int    { return 2 }
