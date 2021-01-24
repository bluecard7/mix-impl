package main

// pg 140
// A byte can hold at least 64 values, at most 100
// meaning 0-63 on binary computer, 0-100 (2 digits) on decimal computer
type Byte interface {
}

// 6 bits per byte (before standard 8 bits), only using last 6 bits
type BinByte byte

type Sign byte // + or -

type BinWord struct {
	Bytes []BinByte // len 6, using [0] as sign for now, 0 == "-", 1 == "+"
	//	sign Sign
}

// field specification (L:R), expressed as 8L + R
// Goes from 0 to 5, where 0 is the sign
func fieldSpec(spec BinByte) (L, R BinByte) {
	L, R = byte(spec/8), spec%8
	return L, R
}

/* Instruction format for words used for instructions
   (sign), A, A, I, F, C
   C: opcode
   F: modification (usually field specification, could be something else)
   (sign)AA: is the address
   I: index specification, specify rI? to modify effective address (can be [0, 6])
*/

type DuoByte struct {
	b    []BinByte // len 2
	sign Sign
}

type RegisterSet struct {
	A                      BinWord // accumulator
	X                      BinWord // extension
	I1, I2, I3, I4, I5, I6 DuoByte // index
	J                      DuoByte // jump address, sign always +
}

type OverflowToggle bool

type ComparisonIndicator struct {
	Less, Equal, Greater bool
}

// Hardware/Architecture
type MIXArch struct {
	Regs RegisterSet
	Mem  []BinWord // 4000 words
}

// also io devices like cards, tapes, disks
func main() {
}
