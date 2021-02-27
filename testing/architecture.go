package main

const (
	WORDSIZE = 5
	BYTESIZE = 6
)

type Word int32
type bitslice struct {
	word       Word
	start, end int // change to byte
}

func (w Word) sign() Word {
	return w & 0x40000000 // first bit used as sign, 1 is negative
}

func (w Word) data() Word {
	return w & 0x3FFFFFFF // last 30 bits used as data, 6 bits/Byte
}

func (w Word) negate() Word {
	return w ^ 0x40000000
}

func (left Word) add(right Word) (sum Word, overflowed bool) {
	return left + right, left < left+right
}

// slice returns the Word in [L:R].
// positive if sign isn't included in the slice.
func (w Word) slice(L, R int) (s bitslice) {
	/*if L == 0 {
		s |= w.sign()
		L = 1
	}
	mask, bytePos := 0x00000000, 0x0000003F
	for i := 5; L <= i; i-- {
		if L <= i || i <= R {
			mask |= bytePos
		}
		bytePos <<= BYTE_SIZE
	}*/
	// or is this just word, L, R?
	// above logic is just masking...
	//return newBitslice(s|(w.data()&mask), L, R)
	return bitslice{w, L, R}
}

// is this more a slice thing?
func (w Word) value(L, R int) (v int32) {
	s := w.slice(L, R)
	v = int32(s.word.data())
	v >>= (WORDSIZE - R) * BYTESIZE
	if 0 < s.word.sign() {
		v = -v
	}
	return v
}

func (s bitslice) len() int {
	if s.start == 0 {
		return s.end - s.start
	}
	return s.end - s.start + 1
}

func bitmask(L, R int) (mask Word) {
	if L == 0 {
		mask = 1 << 30
		L = 1
	}
	var bytePos Word = 0x3F << (WORDSIZE - R) * BYTESIZE
	for i := R; L <= i; i-- {
		mask |= bytePos
		bytePos <<= BYTESIZE
	}
	return mask
}

// TODO:: Deal with signs
// - 1 is used instead if dst.start == 0
// - if dst.start != 0, it's positive
func bitcopy(dst, src bitslice) {
	srcWord := src.word
	shiftAmt := dst.start - src.start
	if shiftAmt < 0 {
		srcWord <<= -shiftAmt * BYTESIZE
	} else {
		srcWord >>= shiftAmt * BYTESIZE
	}
	mask := bitmask(dst.start, dst.end)
	dst.word &= mask ^ mask // zero out portion to be written to
	dst.word |= src.word & mask
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

type Register bitslice

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
func (m *Arch) Cell(address int) *Word {
	return &m.Mem[address]
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
