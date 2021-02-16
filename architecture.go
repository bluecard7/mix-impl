package main

type MIXByte byte
type MIXBytes []MIXByte

// NewByte returns a MIXByte with the value of the given data.
// Since a byte in MIX needs to hold at least 64 distinct values,
// data must be in the range [0, 63].
// If data is negative or > 63, then it will be treated as 0.
func NewByte(data MIXByte) MIXByte {
	if data < 0 || 63 < data {
		data = 0
	}
	return MIXByte(data)
}

const (
	POS_SIGN MIXByte = 0
	NEG_SIGN         = 1

	WORD_SIZE = 6
)

func (b MIXBytes) Sign() MIXBytes {
	return b[:1]
}

func (b MIXBytes) Data() MIXBytes {
	return b[1:]
}

// Slice returns the MIXBytes in [L:R].
// A sign is added if it isn't included in the slice.
func (b MIXBytes) Slice(L, R MIXByte) (s MIXBytes) {
	s = b[L : R+1]
	if 0 < L {
		s = append(MIXBytes{POS_SIGN}, s...)
	}
	return s
}

// Negate takes b and treats the MIXByte at index 0 as a sign.
// It returns a copy with the opposite sign.
func (b MIXBytes) Negate() (negated MIXBytes) {
	opposite := POS_SIGN
	if b[0] == POS_SIGN {
		opposite = NEG_SIGN
	}
	return append(MIXBytes{opposite}, b.Data()...)
}

// Creates MIXBytes of specified size, fills with zeros if
// b is smaller than size
//func (b MIXBytes) Padded(size int) MIXBytes {
//}

// Equals method for slice of MIXBytes
func (left MIXBytes) Equals(right MIXBytes) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func (left MIXBytes) Add(right MIXBytes) (MIXBytes, bool) {
	var overflowed bool
	sum := toNum(left) + toNum(right)
	if sum > 2<<31-1 {
		overflowed = true
	}
	return toMIXBytes(sum, 5), overflowed
}

// NewWord returns MIXBytes(len 6) holding the given data.
// The given sequence (data) will be interpreted as: the sign, then MIX bytes 1, 2, ..., 5.
// If more values are given, they are ignored.
// The given values will be subject to the conditions stated in NewByte, as
// each byte in data is converted into a MIXByte through NewByte.
func NewWord(data ...MIXByte) MIXBytes {
	word := make(MIXBytes, 6)
	for i, datum := range data {
		if 6 < i {
			break
		}
		word[i] = NewByte(datum)
	}
	return word
}

// toNum returns the numeric value of a group of continguous MIX bytes.
// The first MIX byte will be interpreted as a sign (positive or negative).
// (max value is 2^31-1)
func toNum(mixBytes MIXBytes) (value int64) {
	for i := 1; i < len(mixBytes); i++ {
		value <<= 6
		value += int64(mixBytes[i]) & 63
	}
	if mixBytes[0] == NEG_SIGN {
		value = -value
	}
	return value
}

// toMIXBytes converts the given value to a slice of MIX bytes with len size.
// The value will be truncated if it exceeds the allowed capacity.
func toMIXBytes(value int64, size int) MIXBytes {
	mixBytes := make(MIXBytes, size+1)
	if value < 0 {
		mixBytes[0] = NEG_SIGN
		value = -value
	}
	for i := len(mixBytes) - 1; i > 0 && value > 0; i-- {
		mixBytes[i] = NewByte(MIXByte(value & 63))
		value >>= 6
	}
	return mixBytes
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

// Registers use the index 0 as sign and the rest for data
type Register MIXBytes

func (r Register) Sign() MIXBytes {
	return r.Raw().Sign()
}

func (r Register) Data() MIXBytes {
	return r.Raw().Data()
}

func (r Register) Raw() MIXBytes {
	return MIXBytes(r)
}

// MIXArch defines the hardware/architecture elements of the MIX machine
type MIXArch struct {
	R                   []Register
	Mem                 []MIXBytes
	PC                  MIXBytes // program counter
	OverflowToggle      bool
	ComparisonIndicator struct {
		Less, Equal, Greater bool
	}
}

// Cell returns the MIX word at the memory cell at address, indexed by I register.
func (m *MIXArch) Cell(inst Instruction) MIXBytes {
	address, index := toNum(Address(inst)), Index(inst)
	if 0 < index {
		address += toNum(m.R[I1+int(index)-1].Raw())
	}
	return machine.Mem[address]
}

func (m *MIXArch) Exec(inst Instruction) {
	inst.Do(m)
}

func (m *MIXArch) SetComparisons(lt, eq, gt bool) {
	m.ComparisonIndicator.Less = lt
	m.ComparisonIndicator.Equal = eq
	m.ComparisonIndicator.Greater = gt
}

func (m *MIXArch) Comparisons() (bool, bool, bool) {
	return m.ComparisonIndicator.Less, m.ComparisonIndicator.Equal, m.ComparisonIndicator.Greater
}

// NewMachine creates a new instance of MIXArch
func NewMachine() *MIXArch {
	machine := &MIXArch{
		R:   make([]Register, 9),
		Mem: make([]MIXBytes, 4000),
		ComparisonIndicator: struct {
			Less, Equal, Greater bool
		}{},
	}
	machine.R[A] = make(Register, 6)
	machine.R[X] = make(Register, 6)
	machine.R[I1] = make(Register, 3)
	machine.R[I2] = make(Register, 3)
	machine.R[I3] = make(Register, 3)
	machine.R[I4] = make(Register, 3)
	machine.R[I5] = make(Register, 3)
	machine.R[I6] = make(Register, 3)
	machine.R[J] = make(Register, 3)
	for i := range machine.Mem {
		machine.Mem[i] = NewWord()
	}
	return machine
}

// also io devices like cards, tapes, disks
