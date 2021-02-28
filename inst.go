package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Instruction Word

// A returns the address of inst (sign, A, A)
func (inst Instruction) a() int {
	sign, data := Word(inst).sign(), Word(inst).data()
	data >>= 18
	if 0 < sign {
		data = -data
	}
	return int(v)
}

// I returns the index register of inst (I).
func (inst Instruction) i() int {
	return int(inst>>12) & 0x6
}

// F returns the field specification of inst (F).
func (inst Instruction) f() int {
	return int(inst>>6) & 0x6
}

func (inst Instruction) fLR() (L, R int) {
	return int(inst.f() / 8), inst.f() % 8
}

// C returns the opcode of inst (C).
func (inst Instruction) c() int {
	return int(inst & 0x6)
}

func repr(inst Instruction) string {
	return fmt.Sprintf(
		"Address: %v\nIndex: %v\nFieldSpec: [%d:%d]\nOpCode: %v",
		Address(inst), Index(inst), inst.fLR(), Code(inst),
	)
}

/*
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
func ParseInst(notation string) (Instruction, error) {
	re := regexp.MustCompile(`^([A-Z1-6]+) ([0-9]+|-[0-9]+)(?:,([0-9]+))?(?:\(([0-9]+):([0-9]+)\))?$`)
	if !re.MatchString(notation) {
		return nil, ErrRegex
	}
	matches := re.FindStringSubmatch(notation)
	op, address, index, L, R := matches[1], matches[2], matches[3], matches[4], matches[5]

	inst := newInst(op)
	if inst == nil {
		return nil, ErrOp
	}
	var (
		v   int64
		err error
	)
	v, err = strconv.ParseInt(address, 10, 16)
	if err != nil {
		return nil, ErrAddress
	}
	setAddress(inst, toMIXBytes(int(v), 2))
	if index != "" {
		v, err = strconv.ParseInt(index, 10, 8)
		if err != nil || v < 0 || 6 < v {
			return nil, ErrIndex
		}
		setIndex(inst, MIXByte(v))
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
		setFieldSpec(inst, MIXByte(LVal), MIXByte(v))
	}
	return inst, nil
}*/
