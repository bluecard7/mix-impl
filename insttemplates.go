package main

import (
//"regexp"
)

func regIndex(regName string) Word {
	// each character is in ascii range, so b will be a byte.
	// register index constants organized in the same order as string.
	for rI, b := range "A123456X" {
		if regName == string(b) {
			return Word(rI)
		}
	}
	return NoR
}

/*func newInst(instName string) Word {
	if template, ok := nameToTemplate[instName]; ok {
		return template()
	}
	for pattern, template := range patternToTemplate {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(instName)
		if match != nil && regIndex(match[1]); rI != NoR {
			return template(rI)
		}
	}
	return nil
}*/

//CMP[A1-6X] : 56-63
var patternToTemplate = map[string]func(rI Word) Word{
	`^LD([A1-6X])$`:  func(rI Word) Word { return composeInst(0, 0, 5, C_LD+rI) },  // LD_
	`^LD([A1-6X])N$`: func(rI Word) Word { return composeInst(0, 0, 5, C_LDN+rI) }, // LD_N
	`^ST([A1-6X])$`:  func(rI Word) Word { return composeInst(0, 0, 5, C_ST+rI) },  // ST_
	/*`^J([A1-6X])N$`:  func(rI MIXByte) Word { return newJmp(0, 40+rI, rI) },             // J_N
	`^J([A1-6X])Z$`:  func(rI MIXByte) Word { return newJmp(1, 40+rI, rI) },             // J_Z
	`^J([A1-6X])P$`:  func(rI MIXByte) Word { return newJmp(2, 40+rI, rI) },             // J_P
	`^J([A1-6X])NN$`: func(rI MIXByte) Word { return newJmp(3, 40+rI, rI) },             // J_NN
	`^J([A1-6X])NZ$`: func(rI MIXByte) Word { return newJmp(4, 40+rI, rI) },             // J_NZ
	`^J([A1-6X])NP$`: func(rI MIXByte) Word { return newJmp(5, 40+rI, rI) },             // J_NP
	`^INC([A1-6X])$`: func(rI MIXByte) Word { return newAddressTransfer(0, 48+rI, rI) }, // INC_
	`^DEC([A1-6X])$`: func(rI MIXByte) Word { return newAddressTransfer(1, 48+rI, rI) }, // DEC_
	`^ENT([A1-6X])$`: func(rI MIXByte) Word { return newAddressTransfer(2, 48+rI, rI) }, // ENT_
	`^ENN([A1-6X])$`: func(rI MIXByte) Word { return newAddressTransfer(3, 48+rI, rI) }, // ENN_
	`^CMP([A1-6X])$`: func(rI MIXByte) Word { return newCmp(56+rI, rI) },                //CMP_
	*/
}

var nameToTemplate = map[string]func() Word{
	"ADD": func() Word { return composeInst(0, 0, 5, 1) },
	"SUB": func() Word { return composeInst(0, 0, 5, 2) },
	// "MUL"
	// "DIV"
	"STJ": func() Word { return composeInst(0, 0, 2, 32) },
	"STZ": func() Word { return composeInst(0, 0, 5, 33) },

	/*"JMP":  func() Word { return newJmp(0, 39, NoR) },
	"JSJ":  func() Word { return newJmp(1, 39, NoR) },
	"JOV":  func() Word { return newJmp(2, 39, NoR) },
	"JNOV": func() Word { return newJmp(3, 39, NoR) },
	"JL":   func() Word { return newJmp(4, 39, NoR) },
	"JE":   func() Word { return newJmp(5, 39, NoR) },
	"JG":   func() Word { return newJmp(6, 39, NoR) },
	"JGE":  func() Word { return newJmp(7, 39, NoR) },
	"JNE":  func() Word { return newJmp(8, 39, NoR) },
	"JLE":  func() Word { return newJmp(9, 39, NoR) },*/
}
