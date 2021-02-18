package main

/*
	Keep track of state with each "tick"
	either as a diff b/n 2 consecutive states (less mem, would need to compute)
	or just a snapshot (more mem, just save)
*/

type History struct {
}

// startInst is cell of first instruction
func ObservableRun(m *MIXArch, startInst int) *History {
	/*
		time := 0
	*/
}
