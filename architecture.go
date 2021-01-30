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
	POS_SIGN = 0
	NEG_SIGN = 1

	WORD_SIZE = 6
)

// Negate takes b and treats the MIXByte at index 0 as a sign.
// It negates b[0], making positive negative, and vice versa.
func (b MIXBytes) Negate() MIXBytes {
	switch b[0] {
	case POS_SIGN:
		b[0] = NEG_SIGN
	case NEG_SIGN:
		b[0] = POS_SIGN
	}
	return b
}

// Equals method for slice of MIXBytes
func (left MIXBytes) Equals(right MIXBytes) bool {
	if len(left) != len(right) {
		return false
	}
	for i, v := range left {
		if v != right[i] {
			return false
		}
	}
	return true
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
	X         // extension
	I1        // index
	I2
	I3
	I4
	I5
	I6
	J // jump, sign always +
)

// Registers use the index 0 as sign and the rest for data
type Register MIXBytes

func (r Register) Raw() MIXBytes {
	return MIXBytes(r)
}

// MIXArch defines the hardware/architecture elements of the MIX machine
type MIXArch struct {
	R                   []Register
	Mem                 []MIXBytes
	OverflowToggle      bool
	ComparisonIndicator struct {
		Less, Equal, Greater bool
	}
}

// WriteCell (copies/assigns) the given data to the memory cell specified by number.
func (machine MIXArch) WriteCell(address MIXBytes, data MIXBytes) {
	copy(machine.Mem[toNum(address)], data)
}

// ReadCell returns the MIX word at the memory cell at cellNum.
func (machine MIXArch) ReadCell(address MIXBytes) MIXBytes {
	return machine.Mem[toNum(address)]
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
