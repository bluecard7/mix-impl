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
	mixalRe     *regexp.Regexp
}

func NewAssembler() *Assembler {
	return &Assembler{
		definedSyms: make(map[string]int),
		mixalRe:     regexp.MustCompile(`(.+\s)?(.+)\s(.+)`), // code or
	}
}

// should I return Word isntead of int?

const ErrFutureRef = errors.New("symbol: future reference")

// what about local syms?
func (a *Assembler) symbol(s string) (int, error) {
	if len(s) == 0 || 10 < len(s) {
		return 0, errors.New("symbol: 0 or more than 10 characters")
	}
	for _, c := range s {
		if (c < '0' || '9' < c) && (c < 'A' || 'Z' < c) {
			return 0, errors.New("symbol: contains non-digit or non-capital letter")
		}
	}
	v, known := a.definedSym[s]
	if !known {
		return v, ErrFutureRef
	}
	return v, nil
}

func (a *Assembler) number(s string) (int, error) {
	if len(s) == 0 || 10 < len(s) { // c is unicode, but digits are ascii (1 byte)
		return 0, errors.New("number: 0 or more than 10 potential digits")
	}
	for _, c := range s {
		if c < '0' || '9' < c {
			return 0, errors.New("number: contains non-digit")
		}
	}
	v, err := strconv.Atoi(s)
	return v, err
}

func (a *Assembler) literal(s string) (int, error) {
	if len(s) == 0 || 12 < len(s) {
		return 0, errors.New("literal: len needs to be in [1, 11]")
	}
	if s[0] == '=' || s[len(s)-1] == '=' {
		return 0, errors.New("literal: not wrapped with equal")
	}
	// wValue(s[1:len(s)-1])
}

func (a *Assembler) unaryOp(s string) (int, error) {
	unaryOp := s[0]
	if unaryOp == '-' || unaryOp == '+' {
		v, err := a.atom(s[1:])
		if err != nil {
			return 0, err
		}
		if unaryOp == "-" {
			v = -v
		}
		return v, nil
	}
	return 0, errors.New("unaryOp: not a unary op")
}

func (a *Assembler) binaryOp(s string) (int, error) {
	// binOpRe: regexp.MustCompile(`(.+)([+-/*:]|//)(.+)`), // don't know if double slash matches here
	op, i := "", len(s)-1
	for op == "" && i > -1 {
		if c := s[i]; c == '+' || c == '-' || c == '*' || c == ':' || c == '/' {
			op = string(c)
		}
		if op == "/" && 0 < i && s[i-1] == '/' { // checks if op is floor division
			op, i = "//", i-1
		}
		i--
	}
	atomVal, atomErr := a.atom(s[i+1+len(op):])
	if atomErr != nil {
		return 0, atomErr
	}
	exprVal, exprErr := a.expression(s[:i+1])
	if exprErr != nil {
		return 0, exprErr
	}
	switch op {
	case "+":
		return exprVal + atomVal, nil
	case "-":
		return exprVal - atomVal, nil
	case "*":
		return exprVal * atomVal, nil
	case "/":
		return exprVal / atomVal, nil
	case "//":
		return int(exprVal / atomVal), nil
	case ":":
		return 8*exprVal + atomVal, nil
	}
	return 0, errors.New("binaryOp: not an binary operation")
}

func (a *Assembler) Assemble(src io.Reader) ([]string, error) {
	line := bufio.NewScanner(src)
	for line.Scan() {
		if line.Text()[0] == '*' {
			continue
		}
		matches := a.mixalRe.FindStringSubmatch(line.Text())
		if matches == nil {
			return nil, errors.New("not a mixal line")
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
	if "*" == s {
		return a.locCounter, nil
	}
	if v, err := a.number(s); err == nil {
		return v, nil
	}
	if v, err := a.symbol(s); err == nil {
		return v, nil
	}
	return 0, errors.New("atom: not an atom") // more descriptive err? check errors package
}

func (a *Assembler) expression(s string) (int, error) {
	if v, err := a.atom(s); err == nil { // atom
		return v, nil
	}
	if v, err := a.unaryOp(s); err == nil { // +atom, -atom
		return v, nil
	}
	if v, err := a.binaryOp(s); err == nil { // expression binop atom
		return v, nil
	}
	return 0, errors.New("expression: not an expression")
}

func (a *Assembler) a(s string) (int, error) {
	// vacuous
	if s == "" {
		return 0, nil
	}
	// future reference
	if v, err := a.number(s); err == ErrFutureRef {
		// doesn't really return a value... how to deal with this?
		return 0, nil
	}
	// literal constant
	if matches := a.literalRe.FindStringSubmatch(s); matches != nil {
		// would place in a constant record
	}
	// expression
	if v, err := a.expression(s); err == nil {
		return v, nil
	}
	return 0, errors.New("a: not an address")
}

func (a *Assembler) i(s string) (int, error) {
	switch true {
	case s == "":
		return 0, nil
	case s[0] == ',':
		return a.expression(s[1:])
	}
	return 0, errors.New("i: not an index")
}

func (a *Assembler) f(s string) (int, error) {
	switch true {
	case s == "":
		return 5, nil // this is wrong, depends on the normal F spec for the op
	case s[0] == '(' && ')' == s[len(s)-1]:
		return a.expression(s[1 : len(s)-1])
	}
	return 0, errors.New("f: not a field spec")
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
