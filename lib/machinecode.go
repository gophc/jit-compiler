package lib

import (
	"encoding/hex"
	"fmt"
	"github.com/bspaans/jit-compiler/platform"
	"unsafe"
)

type MachineCode []uint8

func (m MachineCode) String() string {
	h := hex.EncodeToString(m)
	result := []rune{' ', ' '}
	for i, c := range h {
		result = append(result, c)
		if i%2 == 1 && i+1 < len(h) {
			result = append(result, ' ')
		}
		if i%16 == 15 && i+1 < len(h) {
			result = append(result, '\n', ' ', ' ')
		}
	}
	return string(result)
}

func (m MachineCode) Execute(debug bool) int {
	mmapFunc, err := platform.MmapCodeSegment(len(m))

	if err != nil {
		fmt.Printf("mmap err: %v", err)
	}
	copy(mmapFunc, m)

	value := nativecall(
		uintptr(unsafe.Pointer(&mmapFunc[0])),
		&callEngine{},
	)

	if debug {
		fmt.Println("\nResult :", value)
		fmt.Printf("Hex    : %x\n", value)
		fmt.Printf("Size   : %d bytes\n\n", len(m))
	}
	return value
}
func (m MachineCode) Add(m2 MachineCode) MachineCode {
	return append(m, m2...)
}
