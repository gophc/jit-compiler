package ir

import (
	"fmt"
	"github.com/bspaans/jit-compiler/asm/x86_64"
	"github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	. "github.com/bspaans/jit-compiler/ir/shared"
	"github.com/bspaans/jit-compiler/ir/statements"
	"github.com/bspaans/jit-compiler/lib"
	"github.com/bspaans/jit-compiler/lib/elf"
)

func Compile(targetArchitecture Architecture, abi ABI, stmts []IR, debug bool) (lib.MachineCode, error) {
	fixedReturn := true
	ctx := NewIRContext(targetArchitecture, abi, func(c *IR_Context) *IR_Context {
		c.Debug = debug
		if fixedReturn && len(stmts) > 0 {
			stmt := stmts[len(stmts)-1]

			for then, ok := stmt.(*statements.IR_AndThen); ok; then, ok = stmt.(*statements.IR_AndThen) {
				stmt = then.Stmt2
			}
			if _, ok := stmt.(*statements.IR_Return); ok {
				c.LastReturn = stmt
			}
		}
		return c
	})
	return CompileWithContext(stmts, ctx)
}

//goland:noinspection GoUnusedExportedFunction
func CompileOrigin(targetArchitecture Architecture, abi ABI, stmts []IR, debug bool) (lib.MachineCode, error) {
	ctx := NewIRContext(targetArchitecture, abi, func(c *IR_Context) *IR_Context {
		c.Debug = debug
		return c
	})
	return CompileWithContext(stmts, ctx)
}

//goland:noinspection GoUnusedExportedFunction
func CompileToBinary(targetArchitecture Architecture, abi ABI, stmts []IR, debug bool, path string) error {
	ctx := NewIRContext(targetArchitecture, abi, func(c *IR_Context) *IR_Context {
		c.Debug = debug
		return c
	})
	ctx.ReturnOperandStack = []lib.Operand{encoding.Rax}
	code, err := CompileWithContext(stmts, ctx)
	if err != nil {
		return err
	}
	return elf.CreateTinyBinary(code, path)
}

//goland:noinspection GoErrorStringFormat
func CompileWithContext(stmts []IR, ctx *IR_Context) (lib.MachineCode, error) {
	debug := ctx.Debug

	var result []uint8
	segments, err := ctx.Architecture.EncodeDataSection(stmts, ctx)
	if err != nil {
		return nil, err
	}
	if debug {
		fmt.Println(segments.String())
	}
	// TODO: do this properly
	ctx.Segments = segments
	dataSection := segments.Encode()

	ctx.InstructionPointer += uint(len(dataSection))
	if debug {
		fmt.Println("_start:")
	}
	if len(dataSection) > 0 {
		// TODO make Architecture dependent
		jmp := x86_64.JMP(encoding.Uint8(len(dataSection)))
		if debug {
			fmt.Printf("0x%x: %s\n", 0, jmp.String())
		}
		result_, err := jmp.Encode()
		if err != nil {
			return nil, err
		}
		result = result_
		if debug {
			fmt.Println(result_)
		}
		result = append(result, dataSection...)
	} else {
		ctx.InstructionPointer = 0
	}
	address := uint(2 + len(dataSection))
	for _, stmt := range stmts {
		code, err := ctx.Architecture.EncodeStatement(stmt, ctx)
		if err != nil {
			return nil, fmt.Errorf("Error encoding %s: %s", stmt, err.Error())
		}
		if debug {
			fmt.Println("\n:: " + stmt.String() + "\n")
		}
		for _, i := range code {
			buf, err := i.Encode()
			if err != nil {
				return nil, fmt.Errorf("Failed to encode %s: %s\n%s", stmt, err.Error(), lib.Instructions(code).String())
			}
			if debug {
				fmt.Printf("0x%x-0x%x 0x%x: %s\n", address, address+uint(len(buf)), ctx.InstructionPointer, i.String())
			}
			address += uint(len(buf))
			if debug {
				fmt.Println(buf)
			}
			result = append(result, buf...)
		}
	}
	if debug {
		fmt.Println()
	}
	return result, nil
}
