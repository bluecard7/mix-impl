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

func composeWord(sign, b1, b2, b3, b4, b5 byte) Word {
	return (sign&1)<<30 | (b1&63)<<24 | (b2&63)<<18 | (b3&63)<<12 | (b4&63)<<6 | (b5 & 63)
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

func bitmask(L, R int) (mask Word) {
	if L == 0 {
		mask = 1 << 30
		L = 1
	}
	var bytePos Word = 0x3F << ((WORDSIZE - R) * BYTESIZE)
	for i := R; L <= i; i-- {
		mask |= bytePos
		bytePos <<= BYTESIZE
	}
	return mask
}

// slice returns the Word in [L:R].
// positive if sign isn't included in the slice.
func (w Word) slice(L, R int) (s *bitslice) {
	w &= bitmask(L, R)
	return &bitslice{w, L, R}
}

func (s *bitslice) value() (v int) {
	v = int(s.word.data() & bitmask(s.start, s.end))
	v >>= (WORDSIZE - s.end) * BYTESIZE
	if 0 < s.word.sign() {
		v = -v
	}
	return v
}

func (s *bitslice) len() int {
	if s.start == 0 {
		return s.end - s.start
	}
	return s.end - s.start + 1
}

func (s1 *bitslice) distance(s2 *bitslice) int {
	s1Start, s2Start := s1.start, s2.start
	if s1Start == 0 {
		s1Start = 1
	}
	if s2Start == 0 {
		s2Start = 1
	}
	return s1Start - s2Start
}

func (dst *bitslice) copy(src *bitslice) {
	srcWord := src.word
	// prefer larger indices if data amt described by src is larger than amt in dst
	if accountLen := src.len() - dst.len(); 0 < accountLen {
		srcWord <<= accountLen * BYTESIZE
	}
	if shiftAmt := dst.distance(src); shiftAmt < 0 {
		srcWord <<= -shiftAmt * BYTESIZE
	} else {
		srcWord >>= shiftAmt * BYTESIZE
	}
	mask := bitmask(dst.start, dst.end)
	dst.word &= mask ^ 0x7FFFFFFF // zero out portion to be written to
	dst.word |= srcWord & mask
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
	R                   []*bitslice
	Mem                 []Word
	PC                  int // program counter
	OverflowToggle      bool
	ComparisonIndicator struct {
		Less, Equal, Greater bool
	}
}

func (m *Arch) Read(address int) Word {
	return m.Mem[address]
}

func (m *Arch) Write(address int, data Word) {
	m.Mem[address] = data
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
		R:   make([]*bitslice, 9),
		Mem: make([]Word, 4000),
		ComparisonIndicator: struct {
			Less, Equal, Greater bool
		}{},
	}
	for i := range machine.R {
		if i == A || i == X {
			machine.R[i] = &bitslice{0, 0, 5}
		} else {
			machine.R[i] = &bitslice{0, 0, 2}
		}
	}
	return machine
}

// also io devices like cards, tapes, disks
