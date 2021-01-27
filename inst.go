package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type Instruction struct {
	Code MIXWord // mix machine code (sign, A, A, I, F,C)
	Time int     // time it takes to execute instruction
	Exec func(machine *MIXArch, inst *Instruction)
}

// baseInstCode returns a MIX word that represents the default information
// format. The defaults are no address, no index register, and the given field range
// and op code c.
func baseInstCode(L, R, c MIXByte) MIXWord {
	f := compressFields(L, R)
	return NewWord(POS_SIGN, 0, 0, 0, f, c)
}

// A returns the address portion of inst.Code (sign, A, A).
// If newA has len 3, then the address is set to it.
func (inst *Instruction) A(newA ...MIXByte) []MIXByte {
	if len(newA) == 3 {
		copy(inst.Code, newA)
	}
	return inst.Code[:3]
}

// I returns the index portion of inst.Code (I).
// If newI has len 1, then the index is set to it.
func (inst *Instruction) I(newI ...MIXByte) MIXByte {
	if len(newI) == 1 {
		inst.Code[3] = NewByte(newI[0])
	}
	return inst.Code[3]
}

// F returns the field specification portion of inst.Code (F).
// It is expressed as (L:R), rather than one number.
// If newF has len 1, then the field specification is set to it.
func (inst *Instruction) F(newF ...MIXByte) (L, R MIXByte) {
	if len(newF) == 2 {
		inst.Code[4] = compressFields(newF[0], newF[1])
	}
	f := inst.Code[4]
	L, R = f/8, f%8
	return L, R
}

// C returns the op portion of inst.Code (C).
func (inst *Instruction) C() MIXByte {
	return inst.Code[5]
}

// compressField expresses the field range as 8*L + R
func compressFields(L, R MIXByte) MIXByte {
	return 8*L + R
}

// When an Instruction is printed, it will be displayed in the format
// [(sign)AA][I][F][C]
func (inst Instruction) String() string {
	return fmt.Sprintf("%d", inst.C()) // fmt.Sprintf("[%d][%d][%d][%d]", inst.A(), inst.I(), inst.F(), inst.C())
}

// parseInst takes a string and translate it according to the
// following notation: "OP ADDRESS, I(F)".
// If I is omitted, then it is treated as 0.
// If F is omitted, then it is treated as the normal F specification
// (This is (0:5) for most operators, could be something else).
func parseInst(notation string) (inst *Instruction) {
	// Groups are: op, address, index, L, R
	re := regexp.MustCompile(`^([A-Z]+) ([0-9]+)(?:,([1-6]))?(?:\(([0-5]):([0-5])\))?$`)
	matches := re.FindStringSubmatch(notation)
	op, address, index, L, R := matches[1], matches[2], matches[3], matches[4], matches[5]
	inst = templates[op]()
	addressVal, _ := strconv.ParseInt(address, 10, 16)
	inst.A(toMIXBytes(addressVal, 2)...)
	if index != "" {
		indexVal, _ := strconv.ParseInt(index, 10, 8)
		inst.I(MIXByte(indexVal))
	}
	if L != "" {
		LVal, _ := strconv.ParseInt(L, 10, 8)
		RVal, _ := strconv.ParseInt(R, 10, 8)
		inst.F(MIXByte(LVal), MIXByte(RVal))
	}
	return inst
}
