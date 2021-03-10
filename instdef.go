package main

const (
	C_ADD           = 1
	C_SUB           = 2
	C_MUL           = 3
	C_DIV           = 4
	C_LD            = 8
	C_LDN           = 16
	C_ST            = 24
	C_ADDR_TRANSFER = 48
	C_CMP           = 56
)

func (m *Arch) Exec(inst Word) {
	switch c := inst.c(); true {
	case c == C_ADD:
		m.Add(inst)
	case c == C_SUB:
		m.Add(inst)
	case c == C_MUL:
		m.Mul(inst)
	case c == C_DIV:
		m.Div(inst)
	case C_LD <= c && c < C_ST:
		m.Load(inst)
	case C_ST <= c && c < C_CMP:
		m.Store(inst)
	case C_CMP <= c:
		m.Compare(inst)
	}
}

func (m *Arch) Add(inst Word) {
	data := m.Read(inst.a()).slice(inst.fLR()).w
	if inst.c() == 2 {
		data = -data
	}
	m.R[A].w, m.OverflowToggle = m.R[A].w.add(data)
}

func (m *Arch) Mul(inst Word) {
	v := m.Read(inst.a()).slice(inst.fLR()).w
	sign, product := int64(1), int64(m.R[A].w)*int64(v)
	if product < 0 {
		sign, product = -1, -product
	}
	m.R[A].w = Word(sign * (product >> 30))
	m.R[X].w = Word(sign * (product & 0x3FFFFFFF))
}

func (m *Arch) Div(inst Word) {
	var q, r Word
	den := int64(m.Read(inst.a()).slice(inst.fLR()).w)
	if den != 0 {
		num := int64(m.R[A].w.data())<<30 | int64(m.R[X].w.data())
		q, r = Word(num/den), Word(num%den)
	}
	if den == 0 || q > (1<<31)-1 {
		m.OverflowToggle = true
		return
	}
	m.R[X].w = m.R[A].w.sign() * r
	m.R[A].w = q
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
/*func (m *Arch) Shift(inst Word) {
	buf, size := int64(m.R[A].w.data()), 5
	_, R := inst.fLR()
	if 1 < R { // shifts rA + rX (data only, not signs)
		// keep sign gap in.w?
		buf = (buf << 32) | m.R[X].w.data()
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

func (m *Arch) Load(inst Word) {
	rI, data := inst.c()-C_LD, m.Read(inst.a())
	if C_LDN <= inst.c() {
		rI, data = inst.c()-C_LDN, -data
	}
	regSlice := m.R[rI]
	if regSlice.w.sign() != data.sign() { // s is either positive bc L != 0 or data's sign
		regSlice.w *= -1
	}
	regSlice.copy(data.slice(inst.fLR()))
}

func (m *Arch) Store(inst Word) {
	regS := Word(0).slice(0, 5) // STZ
	if inst.c() < 33 {
		regS = m.R[inst.c()-C_ST].w.slice(0, 5)
	}
	L, R := inst.fLR()
	cell := m.Read(inst.a())
	buf := Word(0).slice(L, R).copy(regS)
	m.Write(inst.a(), buf.apply(cell))
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

func (m *Arch) Jump(inst Word) {
	_, R := inst.fLR()
	c, address := inst.c(), inst.a()

	// comparison flags and values are gathered
	// here to avoid repeating later.
	lt, eq, gt := m.Comparisons()
	var v Word
	if 39 < inst.c() {
		rI := inst.c() - 40
		v = m.R[rI].w
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
func (m *Arch) AddressTransfer(inst Word) {
	rI := inst.c() - C_ADDR_TRANSFER
	_, R := inst.fLR()
	address := inst.a()
	if R%2 == 1 { // DEC, ENN
		address = -address
	}
	dst := m.R[rI]
	if R < 2 { // INC, DEC
		// but is the slice state stable/correct?
		dst.w, m.OverflowToggle = dst.w.add(address)
	} else { // ENT, ENN
		// does this work for I1-I6, J?
		dst.copy(Word(address).slice(0, 5))
	}
}

func (m *Arch) Compare(inst Word) {
	rI := inst.c() - C_CMP
	L, R := inst.fLR()
	regVal := m.R[rI].w.slice(L, R).w
	cellVal := m.Read(inst.a()).slice(L, R).w
	m.SetComparisons(regVal < cellVal, regVal == cellVal, regVal > cellVal)
}
