package main

import (
	"bufio"
	"fmt"
	"os"
)

func Shell() {
	// need to start a machine
	session := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[MIXAL] ")
		rawLine, err := session.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		line := rawLine[:len(rawLine)-1]
		inst, err := ParseInst(line)
		fmt.Println(inst)
		// if halt, ok := inst.(Halt); ok { return }
	}
}
