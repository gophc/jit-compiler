package lib

import (
	"fmt"
	"strings"
)

type Instruction interface {
	Encode() (MachineCode, error)
	String() string
}

type Instructions []Instruction

func (i Instructions) Encode() (MachineCode, error) {
	var result []uint8
	for _, instr := range i {
		b, err := instr.Encode()
		if err != nil {
			return nil, err
		}
		result = append(result, b...)
	}
	return result, nil
}

func (i Instructions) Add(i2 []Instruction) Instructions {
	return append(i, i2...)
}

func (i Instructions) String() string {
	result := make([]string, len(i))
	for j, instr := range i {
		result[j] = instr.String()
	}
	return strings.Join(result, "\n")
}

func CompileInstruction(instr []Instruction, debug bool) (MachineCode, error) {
	var result []uint8
	address := 0
	for _, i := range instr {
		b, err := i.Encode()
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Printf("0x%x: %s\n", address, i.String())
		}
		address += len(b)
		if debug {
			fmt.Println(b)
		}
		result = append(result, b...)
	}
	return result, nil
}

func InstructionLength(instr Instruction) (int, error) {
	b, err := instr.Encode()
	if err != nil {
		return 0, err
	}
	return len(b), nil

}
