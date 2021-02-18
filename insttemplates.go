package main

import (
	"regexp"
)

func regIndex(regName string) MIXByte {
	// each character is in ascii range, so b will be a byte.
	// register index constants organized in the same order as string.
	for rI, b := range "A123456X" {
		if regName == string(b) {
			return MIXByte(rI)
		}
	}
	return NoR
}

func newInst(instName string) Instruction {
	if template, ok := nameToTemplate[instName]; ok {
		return template()
	}
	for pattern, template := range patternToTemplate {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(instName); match != nil {
			if rI := regIndex(match[1]); rI != NoR {
				return template(rI)
			}
		}
	}
	return nil
}

var patternToTemplate = map[string]func(rI MIXByte) Instruction{
	`^LD([A1-6X])$`:  func(rI MIXByte) Instruction { return newLD(8+rI, rI) },                  // LD_
	`^LD([A1-6X])N$`: func(rI MIXByte) Instruction { return newLD(16+rI, rI) },                 // LD_N
	`^ST([A1-6X])$`:  func(rI MIXByte) Instruction { return newST(24+rI, rI) },                 // ST_
	`^J([A1-6X])N$`:  func(rI MIXByte) Instruction { return newJmp(0, 40+rI, rI) },             // J_N
	`^J([A1-6X])Z$`:  func(rI MIXByte) Instruction { return newJmp(1, 40+rI, rI) },             // J_Z
	`^J([A1-6X])P$`:  func(rI MIXByte) Instruction { return newJmp(2, 40+rI, rI) },             // J_P
	`^J([A1-6X])NN$`: func(rI MIXByte) Instruction { return newJmp(3, 40+rI, rI) },             // J_NN
	`^J([A1-6X])NZ$`: func(rI MIXByte) Instruction { return newJmp(4, 40+rI, rI) },             // J_NZ
	`^J([A1-6X])NP$`: func(rI MIXByte) Instruction { return newJmp(5, 40+rI, rI) },             // J_NP
	`^INC([A1-6X])$`: func(rI MIXByte) Instruction { return newAddressTransfer(0, 48+rI, rI) }, // INC_
	`^DEC([A1-6X])$`: func(rI MIXByte) Instruction { return newAddressTransfer(1, 48+rI, rI) }, // DEC_
	`^ENT([A1-6X])$`: func(rI MIXByte) Instruction { return newAddressTransfer(2, 48+rI, rI) }, // ENT_
	`^ENN([A1-6X])$`: func(rI MIXByte) Instruction { return newAddressTransfer(3, 48+rI, rI) }, // ENN_
	`^CMP([A1-6X])$`: func(rI MIXByte) Instruction { return newCmp(56+rI, rI) },                //CMP_
}

var nameToTemplate = map[string]func() Instruction{
	"ADD": func() Instruction { return newAdd(1) },
	"SUB": func() Instruction { return newAdd(2) },
	// "MUL"
	// "DIV"
	"STJ": func() Instruction { return newST(32, J) },
	"STZ": func() Instruction { return newST(33, A) },

	"JMP":  func() Instruction { return newJmp(0, 39, NoR) },
	"JSJ":  func() Instruction { return newJmp(1, 39, NoR) },
	"JOV":  func() Instruction { return newJmp(2, 39, NoR) },
	"JNOV": func() Instruction { return newJmp(3, 39, NoR) },
	"JL":   func() Instruction { return newJmp(4, 39, NoR) },
	"JE":   func() Instruction { return newJmp(5, 39, NoR) },
	"JG":   func() Instruction { return newJmp(6, 39, NoR) },
	"JGE":  func() Instruction { return newJmp(7, 39, NoR) },
	"JNE":  func() Instruction { return newJmp(8, 39, NoR) },
	"JLE":  func() Instruction { return newJmp(9, 39, NoR) },
}
