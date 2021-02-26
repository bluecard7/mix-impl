package main

type Byte byte
type ByteSequence int64 //

// NewByte returns a Byte with the value of the given data.
// Since a byte in  needs to hold at least 64 distinct values,
// data must be in the range [0, 63].
// If data is negative or > 63, then it will be treated as 0.
func NewByte(data byte) Byte {
	if data < 0 || 63 < data {
		data = 0
	}
	return Byte(data)
}

const (
	POS_SIGN  = 0
	NEG_SIGN  = 1
	WORD_SIZE = 6
)

type Word int

func (b Word) sign() int {
	return b & 0x80000000 // first bit used as sign, 1 is negative
}

func (b Word) data() int {
	return b & 0x3FFFFFFF // last 30 bits used as data, 6 bits/Byte
}

// slice returns the Word in [L:R].
// A sign is added if it isn't included in the slice.

// NOT REALLY A SLICE - its a masked version that has the specified region (L:R)
// Can't really tell which portion is a slice after return...
// is the solution for this is to impl slice headers...like in go?
// Or do I even need to deal with slices?
// for copy, would just need a dst, a src, and a field spec
func (w Word) slice(L, R int) (s Word) {
	if L == 0 {
		s = s | s.sign()
		L = 1
	}
	mask, bytePos := 0x00000000, 0x0000003F
	for i := 5; L <= i; i-- {
		if L <= i || i <= R {
			mask |= bytePos
		}
		bytePos <<= 6
	}
	return s | (w.data() & mask)
}

func (w Word) value(L, R int) (v int) {
	s := w.slice(L, R)
	v := s.data()
	for i := 5; R < i; i-- {
		v >>= 6
	}
	if s < 0 {
		v = -v
	}
	return v
}

// Negate takes b and treats the Byte at index 0 as a sign.
// It returns a copy with the opposite sign.
func (w Word) negate() Word {
	return w ^ 0x80000000
}

func (left Word) Add(right Word) (Word, bool) {
	return left + right, left < left+right
}

const (
	A  = iota // accumulator
	I1        // index
	I2
	I3
	I4
	I5
	I6
	X   // extension
	J   // jump, sign always +
	NoR // No register
)

// Arch defines the hardware/architecture elements of the  machine
type Arch struct {
	R                   []Register
	Mem                 []Word
	PC                  int // program counter
	OverflowToggle      bool
	ComparisonIndicator struct {
		Less, Equal, Greater bool
	}
}

// Cell returns the  word at the memory cell at address, indexed by I register.
func (m *Arch) Cell(address int) Word {
	return machine.Mem[address]
}

func (m *Arch) SetComparisons(lt, eq, gt bool) {
	m.ComparisonIndicator.Less = lt
	m.ComparisonIndicator.Equal = eq
	m.ComparisonIndicator.Greater = gt
}

func (m *Arch) Comparisons() (bool, bool, bool) {
	return m.ComparisonIndicator.Less, m.ComparisonIndicator.Equal, m.ComparisonIndicator.Greater
}

// NewMachine creates a new instance of Arch
func NewMachine() *Arch {
	machine := &Arch{
		R:   make([]Register, 9),
		Mem: make([]Word, 4000),
		ComparisonIndicator: struct {
			Less, Equal, Greater bool
		}{},
	}
	// add func to deal with 2 Byte registers?
	return machine
}

// also io devices like cards, tapes, disks
