package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Instruction struct {
	Code MIXBytes // mix machine code (sign, A, A, I, F,C)
	Time int      // time it takes to execute instruction
	Exec func(machine *MIXArch, inst *Instruction)
}

// baseInstCode returns a MIX word that represents the default information
// format. The defaults are no address, no index register, and the given field range
// and op code c.
func baseInstCode(L, R, c MIXByte) MIXBytes {
	f := compressFields(L, R)
	return NewWord(POS_SIGN, 0, 0, 0, f, c)
}

// A returns the address portion of inst.Code (sign, A, A)
// indexed by the index register at inst.I().
// If newA has len 3, then the address is set to it.
func (inst *Instruction) A(newA ...MIXByte) MIXBytes {
	if len(newA) == 3 {
		copy(inst.Code, newA)
	}
	index := machine.R[I1+int(inst.I())-1].Raw()
	return toMIXBytes(toNum(inst.Code[:3])+toNum(index), 2)
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

var (
	ErrRegex   = errors.New("Invalid characters detected in one or more fields, want \"op address,index(L:R)\" where op is a string and address, index, L, and R are numbers, \",index\" and \"(L:R)\"  are optional")
	ErrOp      = errors.New("Operation is not defined")
	ErrAddress = errors.New("Address not defined, want number")
	ErrIndex   = errors.New("Index not defined, want number in [1, 6]")
	ErrField   = errors.New("Field not defined, want number in [0, 5] and L <= R")
)

// ParseInst takes a string and translate it according to the
// following notation: "OP ADDRESS, I(F)".
// If I is omitted, then it is treated as 0.
// If F is omitted, then it is treated as the normal F specification
// (This is (0:5) for most operators, could be something else).
func ParseInst(notation string) (*Instruction, error) {
	re := regexp.MustCompile(`^([A-Z1-6]+) ([0-9]+|-[0-9]+)(?:,([0-9]+))?(?:\(([0-9]+):([0-9]+)\))?$`)
	if !re.MatchString(notation) {
		return nil, ErrRegex
	}
	matches := re.FindStringSubmatch(notation)
	op, address, index, L, R := matches[1], matches[2], matches[3], matches[4], matches[5]

	template, ok := templates[op]
	if !ok {
		return nil, ErrOp
	}
	inst := template()
	var (
		v   int64
		err error
	)
	v, err = strconv.ParseInt(address, 10, 16)
	if err != nil {
		return nil, ErrAddress
	}
	inst.A(toMIXBytes(v, 2)...)
	if index != "" {
		v, err = strconv.ParseInt(index, 10, 8)
		if err != nil || v < 0 || 6 < v {
			return nil, ErrIndex
		}
		inst.I(MIXByte(v))
	}
	if L != "" {
		v, err = strconv.ParseInt(L, 10, 8)
		if err != nil || v < 0 || 5 < v {
			return nil, ErrField
		}
		LVal := v
		v, err = strconv.ParseInt(R, 10, 8)
		if err != nil || v < 0 || 5 < v || LVal > v {
			return nil, ErrField
		}
		inst.F(MIXByte(LVal), MIXByte(v))
	}
	return inst, nil
}
