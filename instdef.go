package main

const (
	C_ADD           = 1
	C_SUB           = 2
	C_LD            = 8
	C_LDN           = 16
	C_ST            = 24
	C_ADDR_TRANSFER = 48
	C_CMP           = 56
)

func (m *Arch) Exec(inst Instruction) {
	switch c := inst.c(); true {
	case c == C_ADD:
		m.Add(inst)
	case c == C_SUB:
		m.Add(inst)
	case C_LD <= c || c < C_ST:
		m.Load(inst)
	case C_ST <= c:
		m.Store(inst)
	case C_CMP <= c:
		m.Compare(inst)
	}
}

func (m *Arch) Add(inst Instruction) {
	data := m.Read(inst.a()).slice(inst.fLR()).word
	if inst.c() == 2 {
		data = data.negate()
	}
	m.R[A].word, m.OverflowToggle = m.R[A].word.add(data)
}

/*type Convert struct {
	fields MIXBytes
}
func newConv(R MIXByte) *Convert {
	return &Convert{defaultFields(0, R, 5)}
}
// TODO:: conversions NUM and CHAR*/

//func newShift(R MIXByte) *Shift {
//	return &Shift{defaultFields(0, R, 6)}
/*func (m *Arch) Shift(inst Instruction) {
	buf, size := int64(m.R[A].word.data()), 5
	_, R := inst.fLR()
	if 1 < R { // shifts rA + rX (data only, not signs)
		// keep sign gap in word?
		buf = (buf << 32) | m.R[X].word.data()
		size += 5
	}
	var (
		shiftAmt   = inst.a() % size
		removed    int64 // make(MIXBytes, shiftAmt)
		vacantMask Word  // needs to be int64 as well, just combine 2 masks for now
	)
	if R%2 == 0 { // left shift
		// vacant = rData[size-shiftAmt:]
		vacantMask = bitmask(size-shiftAmt, size-1)
		// copy(removed, rData[:shiftAmt])
		removed |= buf & bitmask(0, shiftAmt-1)
		// copy(rData, rData[shiftAmt:])
		// buf & (mask ^ 0x7FFFFFFF) | (buf & mask)
	} else { // right shift
		// vacant = rData[:shiftAmt]
		vacantMask = bitmask(0, shiftAmt-1)
		// copy(removed, rData[size-shiftAmt:])
		removed |= buf & bitmask(size-shiftAmt, size-1)
		// copy(rData[shiftAmt:], rData[:size-shiftAmt])
	}
	if 3 < R { // circular
		//copy(vacant, removed)
	} else {
		//copy(vacant, make(MIXBytes, len(vacant)))
	}
	//copy(m.R[A].Data(), rData)
	//copy(m.R[X].Data(), rData[5:]) // think nop if rX wasn't included in shift
}*/

/*func newMove(F MIXByte) *Move {
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
}*/

func (m *Arch) Load(inst Instruction) {
	rI := inst.c() - C_LD
	data := m.Read(inst.a())
	if C_LDN <= inst.c() {
		rI = inst.c() - C_LDN
		data = data.negate()
	}
	regSlice := m.R[rI]
	if regSlice.word.sign() != data.sign() { // s is either positive bc L != 0 or data's sign
		regSlice.word.negate()
	}
	regSlice.copy(data.slice(inst.fLR()))
	// above should've changed m.R[rI]
}

/*func newST(c, rI MIXByte) *Store {
	st := &Store{
		fields: defaultFields(0, 5, c),
		rI:     rI,
	}
	if c == 32 {
		setFieldSpec(st, 0, 2)
	}
	return st
}*/
func (m *Arch) Store(inst Instruction) {
	var rI Word
	switch inst.c() {
	case 32:
		rI = J
	case 33:
		rI = A
	default:
		rI = inst.c() - C_ST
	}
	var regSlice *bitslice
	switch true {
	case I1 <= rI && rI <= I6:
		regWord := m.R[rI].word
		regWord = regWord.sign() | regWord.data()>>18
		regSlice = regWord.slice(0, 5)
	case inst.c() < 33:
		regSlice = m.R[rI]
	default: //STZ
		regSlice = Word(0).slice(0, 5)
	}
	cellSlice := m.Read(inst.a()).slice(inst.fLR())
	// is sign accounted for?
	cellSlice.copy(regSlice)
	m.Write(inst.a(), cellSlice.word)
}

/*type IO struct {
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

	return new(Snapshot)
}*/

func (m *Arch) Jump(inst Instruction) {
	_, R := inst.fLR()
	c, address := inst.c(), inst.a()

	// comparison flags and values are gathered
	// here to avoid repeating later.
	lt, eq, gt := m.Comparisons()
	var v int
	if 39 < inst.c() {
		rI := inst.c() - 40
		v = m.R[rI].value()
	}

	// Jumping seems to consist of writing to
	// rJ and PC.
	_setJmp := func() {
		m.R[J].copy(Word(address).slice(0, 5))
		m.PC = address
	}

	switch true {
	case c == 39 && R == 0: // JMP
		_setJmp()
	case c == 39 && R == 1: // JSJ
		m.PC = address
	case c == 39 && R == 2: // JOV
		if m.OverflowToggle {
			_setJmp()
		}
		m.OverflowToggle = false
	case c == 39 && R == 3: // JNOV
		if !m.OverflowToggle {
			_setJmp()
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
}

//func newAddressTransfer(R, c, rI MIXByte) *AddressTransfer {
func (m *Arch) AddressTransfer(inst Instruction) {
	rI := inst.c() - C_ADDR_TRANSFER
	_, R := inst.fLR()
	address := inst.a()
	if R%2 == 1 { // DEC, ENN
		address = address.negate()
	}
	dst := m.R[rI]
	if R < 2 { // INC, DEC
		// but is the slice state stable/correct?
		dst.word, m.OverflowToggle = dst.word.add(address)
	} else { // ENT, ENN
		// does this work for I1-I6, J?
		dst.copy(Word(address).slice(0, 5))
	}
}

func (m *Arch) Compare(inst Instruction) {
	rI := inst.c() - C_CMP
	L, R := inst.fLR()
	regSlice := m.R[rI].word.slice(L, R)
	cellSlice := m.Read(inst.a()).slice(L, R)
	regVal, cellVal := regSlice.value(), cellSlice.value()
	m.SetComparisons(regVal < cellVal, regVal == cellVal, regVal > cellVal)
}
