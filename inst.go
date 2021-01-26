package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type Instruction struct {
	Code MIXWord // sign, A, A, I, F,C
	Time int     // time it takes to execute instruction
	Exec func(arch MIXArch)
}

// baseInstCode
func baseInstCode(L, R, c MIXByte) MIXWord {
	f := compressFieldRange(L, R)
	return NewWord(POS_SIGN, 0, 0, 0, f, c)
}

func (inst *Instruction) A(newA ...byte) MIXDuoByte {
	if len(newA) == 3 {
		copy(inst.Code, NewDuoByte(newA))
	}
	return inst.Code[:3]
}

func (inst *Instruction) I(newI ...byte) MIXByte {
	if len(newI) > 0 {
		inst.Code[3] = NewByte(newC[0])
	}
	return inst.Code[3]
}

// F
func (inst *Instruction) F(newF ...byte) MIXByte {
	if len(newF) > 0 {
		inst.Code[4] = compressFieldRange(newF[0], newF[1])
	}
	return inst.Code[4]
}

// C
func (inst *Instruction) C(newC ...byte) MIXByte {
	if len(newC) > 0 {
		inst.Code[5] = NewByte(newC[0])
	}
	return inst.Code[5]
}

// FieldRange
func (inst *Instruction) FieldRange() (L, R MIXByte) {
	L, R = inst.F()/8, inst.F()%8
	return L, R
}

// compressFieldRange expresses the field range as 8*L + R
func compressFieldRange(L, R MIXByte) MIXByte {
	return 8*L + R
}

// When an Instruction is printed, it will be displayed in the format
// [(sign)AA][I][F][C]
func (inst Instruction) String() string {
	return fmt.Sprintf("[%d][%d][%d][%d]", inst.A, inst.I, inst.F, inst.C)
}

// parseInst takes a string and translate it according to the
// following notation: "OP ADDRESS, I(F)".
// If I is omitted, then it is treated as 0.
// If F is omitted, then it is treated as the normal F specification
// (This is (0:5) for most operators, could be something else).
func parseInst(notation string) (inst *Instruction) {
	re := regexp.MustCompile(`^(?P<op>[A-Z]+) (?P<address>[0-9]+)(,(?P<index>[1-6]))?(\((?P<L>[0-5]):(?P<R>[0-5])\))?$`)
	matches := re.FindSubmatch([]byte(notation))
	getMatchByName := func(re *regexp.Regexp, name string) string {
		if i := re.SubexpIndex(name); -1 < i {
			return string(matches[i])
		}
		return ""
	}
	inst = instTemplates[getMatchByName("op")]()
	addressVal, _ := strconv.ParseInt(getMatchByName("address"), 10, 16)
	inst.A(toMIXBytes(addressVal, 2)...)
	if index := getMatchByName("index"); index != "" {
		indexVal, _ := strconv.ParseInt(index, 10, 8)
		inst.I(byte(indexVal))
	}
	if L, R := getMatchByName("L"), getMatchByName("R"); L != "" {
		LVal, _ := strconv.ParseInt(L, 10, 8)
		RVal, _ := strconv.ParseInt(R, 10, 8)
		inst.F(byte(LVal), byte(RVal))
	}
	return inst
}

var instTemplates = map[string]func() Instruction{
	"LDA": func() *Instruction {
		return &Instruction{
			Code: baseInstCode(0, 5, 8),
			Exec: func(arch *MIXArch, inst *Instruction) {
				contents := arch.Mem[toNumValue(inst.A()...)]
				L, R := inst.FieldRange()
				arch.Regs.A[0] = POS_SIGN
				if L == 0 {
					arch.Regs.A[0] = contents[0]
					L = 1
				}
				partial := contents[L : R+1]
				copy(arch.Regs.A[6-len(partial):], partial)
			},
		}
	},
}
