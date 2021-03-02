package main

const (
	WORDSIZE = 5
	BYTESIZE = 6
)

type Word int32
type bitslice struct {
	w, len Word // Not really safe cause weird if negative
}

// to negate, just do -composeWord(...)
func composeWord(b1, b2, b3, b4, b5 Word) (w Word) {
	w = (b1&63)<<24 | (b2&63)<<18 | (b3&63)<<12 | (b4&63)<<6 | (b5 & 63)
	return w
}

func (w Word) sign() Word {
	if w < 0 {
		return -1
	}
	return 1
}

func (w Word) data() (v Word) {
	v = w
	if w < 0 {
		v = -w
	}
	return v & 0x3FFFFFFF // last 30 bits used as data, 6 bits/Byte
}

func (left Word) add(right Word) (sum Word, overflowed bool) {
	return left + right, left < left+right
}

func bitmask(L, R Word) (mask Word) {
	if L == 0 {
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
func (w Word) slice(L, R Word) (s *bitslice) {
	v := w.data() & bitmask(L, R)
	v >>= ((WORDSIZE - R) * BYTESIZE)
	if L == 0 {
		v *= w.sign()
		L = 1
	}
	return &bitslice{w, R - L + 1}
}

func (dst *bitslice) copy(src *bitslice) {
	srcWord, copyAmt := src.w, dst.len
	if src.len < dst.len {
		copyAmt = src.len
	}
	mask := bitmask(WORDSIZE-compyAmt+1, WORDSIZE)
	dst.w &= mask ^ 0x7FFFFFFF // zero out portion to be written to
	dst.w |= srcWord & mask
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
	PC                  Word // program counter
	OverflowToggle      bool
	ComparisonIndicator struct {
		Less, Equal, Greater bool
	}
}

func (m *Arch) Read(address Word) Word {
	return m.Mem[address]
}

func (m *Arch) Write(address, data Word) {
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
			machine.R[i] = Word(0).slice(0, 5)
		} else {
			machine.R[i] = Word(0).slice(0, 2)
		}
	}
	return machine
}

// also io devices like cards, tapes, disks
