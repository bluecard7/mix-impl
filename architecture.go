package main

type MIXByte byte

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
)

// A word in MIX is 5 bytes and a sign.
// MIXWord uses a slice of 6 MIXBytes with index 0 for sign.
type MIXWord []MIXByte

// NewWord returns a MIXWord holding the given data.
// The given sequence (data) will be interpreted as: the sign, then MIX bytes 1, 2, ..., 5.
// If more values are given, they are ignored.
// The given values will be subject to the conditions stated in NewByte, as
// each byte in data is converted into a MIXByte through NewByte.
func NewWord(data ...MIXByte) MIXWord {
	word := make([]MIXByte, 6)
	for i, datum := range data {
		if 6 < i {
			break
		}
		word[i] = NewByte(datum)
	}
	return word
}

// A DuoByte is a term used to describe 2 continguous MIX bytes.
// MIXDuoByte uses a slice of 3 MIXBytes with index 0 for sign.
type MIXDuoByte []MIXByte

// NewDuoByte returns a MIXDuoByte holding the given data.
// The given sequence (data) will be interpreted as: the sign, then MIX bytes 1 and 2.
// If more values are given, they are ignored.
// The given values will be subject to the conditions stated in NewByte, as
// each byte in data is converted into a MIXByte through NewByte.
func NewDuoByte(data ...MIXByte) MIXDuoByte {
	duoByte := make([]MIXByte, 3)
	for i, datum := range data {
		if 3 < i {
			break
		}
		duoByte[i] = NewByte(datum)
	}
	return duoByte
}

// toNum returns the numeric value of a group of continguous MIX bytes.
// The first MIX byte will be interpreted as a sign (positive or negative).
// (Using int as return value since a MIX word has 30 bits.)
func toNum(mixBytes ...MIXByte) int {
	value := 0
	for i := 1; i < len(mixBytes); i++ {
		value <<= 6
		value += int(mixBytes[i]) & 63
	}
	if mixBytes[0] == NEG_SIGN {
		value = -value
	}
	return value
}

// toMIXBytes converts the given value to a slice of MIX bytes with len size.
// The value will be truncated if it exceeds the allowed capacity.
func toMIXBytes(value int64, size int) []MIXByte {
	mixBytes := make([]MIXByte, 1, size+1)
	if value < 0 {
		mixBytes[0] = NEG_SIGN
		value = -value
	}
	for value > 0 && len(mixBytes) < size {
		mixBytes = append(mixBytes, NewByte(MIXByte(value&63)))
		value >>= 6
	}
	return mixBytes
}

type RegisterSet struct {
	A                      MIXWord    // accumulator
	X                      MIXWord    // extension
	I1, I2, I3, I4, I5, I6 MIXDuoByte // index
	J                      MIXDuoByte // jump address, sign always +
}

type ComparisonIndicator struct {
	Less, Equal, Greater bool
}

// MIXArch defines the hardware/architecture elements of the MIX machine
type MIXArch struct {
	Regs           RegisterSet
	Mem            []MIXWord
	OverflowToggle bool
	Compare        ComparisonIndicator
}

// WriteCell (copies/assigns) the given data to the memory cell specified by number.
// Consider using DuoByte for cell, cause that's how it works in the machine
// If so needs function to translate values across continguous MIX bytes
func (machine MIXArch) WriteCell(address MIXDuoByte, data MIXWord) {
	copy(machine.Mem[toNum(address...)], data)
}

// ReadCell returns the MIX word at the memory cell at cellNum.
func (machine MIXArch) ReadCell(address MIXDuoByte) MIXWord {
	return machine.Mem[toNum(address...)]
}

// NewMachine creates a new instance of MIXArch
func NewMachine() *MIXArch {
	mem := make([]MIXWord, 4000)
	for i := range mem {
		mem[i] = NewWord()
	}
	return &MIXArch{
		Regs: RegisterSet{
			A:  NewWord(),
			X:  NewWord(),
			I1: NewDuoByte(),
			I2: NewDuoByte(),
			I3: NewDuoByte(),
			I4: NewDuoByte(),
			I5: NewDuoByte(),
			I6: NewDuoByte(),
			J:  NewDuoByte(),
		},
		Mem:     mem,
		Compare: ComparisonIndicator{},
	}
}

// also io devices like cards, tapes, disks
