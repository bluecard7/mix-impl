package main

import (
	"bufio"
	"io"
)

/*
	Keep track of state with each "tick"
	either as a diff b/n 2 consecutive states (less mem, would need to compute)
	or just a snapshot (more mem, just save)
*/

// Keeps track of registers, memory, and devices
// affected by an instruction.
// (indices for R and Mem, and unit # for devices)
// data is interpreted as a stream containing
// changed register, memory, and device data in
// that order.
type Snapshot struct {
	RI, MemI, DeviceI []int
	Data              MIXBytes
}

// Maybe a better approach would be to pair a label
// with each index so it wouldn't matter what order they're added.
// PROFILE: How much overhead due to defer?
func (s *Snapshot) includesR(i int, data Register) {
	s.RI = append(s.RI, i)
	s.Data = append(s.Data, data...)
}

func (s *Snapshot) includesCell(i int, data MIXBytes) {
	s.MemI = append(s.MemI, i)
	s.Data = append(s.Data, data...)
}

func (s *Snapshot) includesDevice(i int, data MIXBytes) {
	s.DeviceI = append(s.DeviceI, i)
	s.Data = append(s.Data, data...)
}

type History []*Snapshot

// really just interpreting
func compile(src io.Reader) ([]Instruction, error) {
	program := make([]Instruction, 0, 20)
	line := bufio.NewScanner(src)
	for line.Scan() {
		inst, err := ParseInst(line.Text())
		if err != nil {
			return nil, err
		}
		program = append(program, inst)
	}
	return program, line.Err()
}

// Note: Instructions aren't actually loaded in machine memory
func Run(m *MIXArch, src io.Reader) (History, error) {
	program, err := compile(src)
	if err != nil {
		return nil, err
	}
	history := make(History, len(program))
	for pos, start := 0, m.PC; pos < len(program); {
		inst := program[m.PC-start]
		effect := m.Exec(inst)
		history = append(history, effect)
	}
	return history, nil
}
