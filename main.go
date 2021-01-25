package main

type MIXByte byte

// NewByte returns a MIXByte with the value of the given data.
// Since a byte in MIX needs to hold at least 64 distinct values,
// data must be in the range [0, 63].
// If data is negative or > 63, then it will be treated as 0.
func NewByte(data byte) MIXByte {
	if data < 0 || 63 < data {
		data = 0
	}
	return MIXByte(data)
}

const (
	POSITIVE = 0
	NEGATIVE = 1
)

// A word in MIX is 5 bytes and a sign.
// MIXWord uses a slice of 6 MIXBytes with index 0 for sign.
type MIXWord []MIXByte

// how to handle values larger than 63 across MIX bytes?
// -> responsibility of some other function (something that breaks a value
// that doesn't fit in a MIXByte to the correct []MIXByte)

// NewWord returns a MIXWord holding the given data.
// The given sequence (data) will be interpreted as: the sign, then MIX bytes 1, 2, ..., 5.
// If more values are given, they are ignored.
// The given values will be subject to the conditions stated in NewByte, as
// each byte in data is converted into a MIXByte through NewByte.
func NewWord(data ...byte) MIXWord {
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
func NewDuoByte(data ...byte) MIXDuoByte {
	duoByte := make([]MIXByte, 3)
	for i, datum := range data {
		if 3 < i {
			break
		}
		duoByte[i] = NewByte(datum)
	}
	return duoByte
}

/* Instruction format for words used for instructions
   (sign), A, A, I, F, C
   C: opcode
   F: modification (usually field specification, could be something else)
   (sign)AA: is the address
   I: index specification, specify rI? to modify effective address (can be [0, 6])
*/

type RegisterSet struct {
	A                      MIXWord    // accumulator
	X                      MIXWord    // extension
	I1, I2, I3, I4, I5, I6 MIXDuoByte // index
	J                      MIXDuoByte // jump address, sign always +
}

type ComparisonIndicator struct {
	Less, Equal, Greater bool
}

// Hardware/Architecture
type MIXArch struct {
	Regs           RegisterSet
	Mem            []MIXWord
	OverflowToggle bool
	Compare        ComparisonIndicator
}

// WriteCell (copies/assigns) the given data to the memory cell specified by number.
// Consider using DuoByte for cell, cause that's how it works in the machine
// If so needs function to translate values across continguous MIX bytes
func (machine MIXArch) WriteCell(cellNum uint16, data MIXWord) {
	// copy or assign?
	machine.Mem[cellNum] = data
}

// ReadCell returns the MIX word at the memory cell at cellNum.
func (machine MIXArch) ReadCell(cellNum uint16) MIXWord {
	return machine.Mem[cellNum]
}

func NewMachine() MIXArch {
	return MIXArch{
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
		Mem: make([]MIXWord, 4000),
	}
}

// field specification (L:R), expressed as 8L + R
// Goes from 0 to 5, where 0 is the sign
func fieldSpec(spec MIXByte) (L, R MIXByte) {
	L, R = spec/8, spec%8
	return L, R
}

// also io devices like cards, tapes, disks
func main() {
}
