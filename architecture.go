package main

import "fmt"

const (
	WORDSIZE = 5
	BYTESIZE = 6
)

type Word int32
type bitslice struct {
	w, start, len Word // start = actual start (w/ sign), len = length of data
}

// to negate, just do -composeWord(...)
func composeWord(b1, b2, b3, b4, b5 Word) (w Word) {
	w = (b1&63)<<24 | (b2&63)<<18 | (b3&63)<<12 | (b4&63)<<6 | (b5 & 63)
	return w
}

func (w Word) view() string {
	sign, data := "+", w.data()
	if w < 0 {
		sign = "-"
	}
	return fmt.Sprintf(
		"%s %v %v %v %v %v",
		sign, data>>24, data>>18&63, data>>12&63, data>>6&63, data&63,
	)
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
		v = -v
	}
	return v & 0x3FFFFFFF // last 30 bits used as data, 6 bits/Byte
}

func (left Word) add(right Word) (sum Word, overflowed bool) {
	return left + right, left < left+right
}

// really only bitmask for data
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
	v := w.data() & bitmask(L, R) >> ((WORDSIZE - R) * BYTESIZE)
	dataStart := L
	if L == 0 {
		v *= w.sign()
		dataStart = 1
	}
	return &bitslice{v, L, R - dataStart + 1}
}

func (dst *bitslice) copy(src *bitslice) {
	srcData, copyAmt := src.w.data(), src.len
	if dst.len < copyAmt {
		copyAmt = dst.len
	}
	mask := bitmask(WORDSIZE-copyAmt+1, WORDSIZE)
	data := dst.w.data()&(mask^0x7FFFFFFF) | (srcData & mask)
	if src.start == 0 {
		dst.w = data * src.w.sign()
	} else {
		dst.w = data * dst.w.sign()
	}
}

func (b *bitslice) apply(w Word) Word {
	sign, L := w.sign(), b.start
	if L == 0 {
		sign = b.w.sign()
		L = 1
	}
	shiftAmt := (WORDSIZE - b.len + 1 - L) * BYTESIZE
	mask, toPos := bitmask(L, L+b.len-1), b.w.data()<<shiftAmt
	return sign * (w.data()&(mask^0x3FFFFFFF) | toPos)
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
