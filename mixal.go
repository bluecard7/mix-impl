package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

type Assembler struct {
	locCounter  int
	definedSyms map[string]int
	symRe       *regexp.Regexp
	numRe       *regexp.Regexp
	unaryOpRe   *regexp.Regexp
	binOpRe     *regexp.Regexp
	literalRe   *regexp.Regexp
	mixalRe     *regexp.Regexp
}

func NewAssembler() *Assembler {
	return &Assembler{
		definedSyms: make(map[string]int),
		symRe:       regexp.MustCompile(`[A-Z0-9]{1,10}|[0-9][HFB]`),
		numRe:       regexp.MustCompile(`[0-9]{1,10}`),
		unaryOpRe:   regexp.MustCompile(`([+-])(.+)`),
		binOpRe:     regexp.MustCompile(`(.+)([+-/*:]|//)(.+)`), // don't know if double slash matches here
		literalRe:   regexp.MustCompile(`=(.+)=`),
		mixalRe:     regexp.MustCompile(`(.+\s)?(.+)\s(.+)`), // code or
	}
}

func (a *Assembler) Assemble(src io.Reader) ([]string, error) {
	line := bufio.NewScanner(src)
	for line.Scan() {
		if line.Text()[0] == '*' {
			continue
		}
		matches := a.mixalRe.FindStringSubmatch(line.Text())
		if matches == nil {
			return nil, errors.New("Line is not mixal")
		}
		sym, op, address := matches[1], matches[2], matches[3]
		fmt.Println(sym, op, address)
		switch op {
		case "EQU":
			v, err := wValue(address)
			if err != nil {
				return nil, err
			}
			a.definedSyms[sym] = v
		case "ORIG":
			if _, know := a.definedSym[sym]; sym != "" && !know {
				a.definedSym[sym] = a.locCounter
			} else {
				return nil, err
			}
			v, err := wValue(address)
			if err != nil {
				return nil, err
			}
			a.locCounter = v
		case "CON":
			v, err := wValue(address)
			if err != nil {
				return nil, err
			}
			// m.Mem[a.locCounter] = v
			a.locCounter++
		case "ALF":
			// assemble address[:5] as alphanumeric char MIX word
			// m.Mem[a.locCounter] = v
			a.locCounter++
		case "END":
			v, err := wValue(address)
			if err != nil {
				return nil, err
			}
		// process each recorded constant as CON
		// also would have something similar for unknown syms
		// if sym != "" ...
		// m.Mem[a.locCounter] = v
		default:
			// ParseInst logic here
		}
	}
	return nil, line.Err()
}

// does it add to internal instruction slice in assembler?
func (a *Assembler) atom(s string) (int, error) {
	switch true {
	case "*" == s:
		return a.locCounter, nil
	case a.numRe.MatchString(s):
		return strconv.Atoi(s)
	case a.symRe.MatchString(s):
		if v, ok := a.definedSyms[s]; ok {
			return v, nil
		}
	}
	return 0, errors.New("Not an atom") // more descriptive err? check errors package
}

func (a *Assembler) expression(s string) (int, error) {
	// atom
	if v, err := a.atom(s); err == nil {
		return v, nil
	}
	// +atom, -atom
	if matches := a.unaryOpRe.FindStringSubmatch(s); matches != nil {
		op, atomStr := matches[1], matches[2]
		if v, err := a.atom(atomStr); err == nil {
			if op == "-" {
				v = -v
			}
			return v, nil
		}
	}
	// expression binop atom
	if matches := a.binOpRe.FindStringSubmatch(s); matches != nil {
		exprStr, op, atomStr := matches[1], matches[2], matches[3]
		if v1, err1 := a.atom(atomStr); err1 == nil {
			if v2, err2 := a.expression(exprStr); err2 == nil {
				switch op {
				case "+":
					return v1 + v2, nil
				case "-":
					return v1 - v2, nil
				case "*":
					return v1 * v2, nil
				case "/":
					return v1 / v2, nil
				case "//":
					return int(v1 / v2), nil
				case ":":
					return 8*v1 + v2, nil
				}
			}
		}
	}
	return 0, errors.New("Not an expression")
}

func (a *Assembler) a(s string) (int, error) {
	// vacuous
	if s == "" {
		return 0, nil
	}
	// future reference
	if a.symRe.MatchString(s) {
		if v, ok := a.definedSyms[s]; !ok {
			return v, nil
		}
	}
	// literal constant
	if matches := a.literalRe.FindStringSubmatch(s); matches != nil {
		// would place in a constant record
	}
	// expression
	if v, err := a.expression(s); err == nil {
		return v, nil
	}
	return 0, errors.New("Not an address")
}

func (a *Assembler) i(s string) (int, error) {
	switch true {
	case s == "":
		return 0, nil
	case s[0] == ',':
		return a.expression(s)
	}
	return 0, errors.New("Not an index")
}

func (a *Assembler) f(s string) (int, error) {
	switch true {
	case s == "":
		return 5, nil // really the normal F spec for the op
	case s[0] == '(' && s[len(s)-1] == ')':
		return a.expression(s[1 : len(s)-1])
	}
	return 0, errors.New("Not a field spec")
}

func (a *Assembler) wValue() {
	// 1. expression + f(), if f() is empty, means [0, 5]
	// 2. Expr(Field), Expr(Field), ...
	// w := Word(0)
	// e, err := expression(Expr)
	// if err != nil {...}
	// w.slice(f).copy(e.slice(0, 5))... basically ST W(f) with e
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
} */
