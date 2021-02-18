package main

type MIXDevice []MIXBytes // slice of MIX words
type MIXDevices []MIXDevice

const (
	TAPE0 = iota // 100 words per tape
	TAPE1
	TAPE2
	TAPE3
	TAPE4
	TAPE5
	TAPE6
	TAPE7
	DISK0 // 100 words per tape
	DISK1
	DISK2
	DISK3
	DISK4
	DISK5
	DISK6
	DISK7
	CARD_READER  // 16 words
	CARD_PUNCHER // 16 words
	LINE_PRINTER // 24 words
	TERMINAL     // 14 words
	PAPER_TAPE   // 14 words
)

// Need:
// character code mapping (could just map sections of ascii to MIX chars)
// devices also have position / last written index
// model device operations with "parallel" work relative to machine (repr time with counter)
// readiness
func peripherals() MIXDevices {
	newDevice := func(blockSize int) MIXDevice {
		device := make(MIXDevice, blockSize)
		for i := range device {
			device[i] = NewWord()
		}
		return device
	}

	devices := make(MIXDevices, 20)
	for i := TAPE0; i <= TAPE7; i++ {
		devices[i] = newDevice(100)
	}
	for i := DISK0; i <= DISK7; i++ {
		devices[i] = newDevice(100)
	}
	devices[CARD_READER] = newDevice(16)
	devices[CARD_PUNCHER] = newDevice(16)
	devices[LINE_PRINTER] = newDevice(24)
	devices[TERMINAL] = newDevice(14)
	devices[PAPER_TAPE] = newDevice(14)
	return devices
}
