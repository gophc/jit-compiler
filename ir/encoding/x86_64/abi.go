package x86_64

import (
	"fmt"

	"github.com/bspaans/jit-compiler/asm/x86_64"
	"github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	. "github.com/bspaans/jit-compiler/ir/shared"
	"github.com/bspaans/jit-compiler/lib"
)

// ABI_AMDSystemV The calling convention of the System V AMD64 ABI is followed on Solaris,
// Linux, FreeBSD, macOS, and is the de facto standard among Unix and Unix-like
// operating systems. The first six integer or pointer arguments are passed in
// registers RDI, RSI, RDX, RCX, R8, R9 (R10 is used as a static chain pointer
// in case of nested functions[25]:21), while XMM0, XMM1, XMM2, XMM3, XMM4,
// XMM5, XMM6 and XMM7 are used for the first floating point arguments.
//
//goland:noinspection GoSnakeCaseUsage
type ABI_AMDSystemV struct {
	intTargets   []*encoding.Register
	floatTargets []*encoding.Register
}

//goland:noinspection GoSnakeCaseUsage
func NewABI_AMDSystemV() *ABI_AMDSystemV {
	return &ABI_AMDSystemV{
		intTargets:   []*encoding.Register{encoding.Rdi, encoding.Rsi, encoding.Rdx, encoding.Rcx, encoding.R10, encoding.R8, encoding.R9},
		floatTargets: []*encoding.Register{encoding.Xmm0, encoding.Xmm1, encoding.Xmm2, encoding.Xmm3, encoding.Xmm4, encoding.Xmm5},
	}
}
func (a *ABI_AMDSystemV) GetRegistersForArgs(args []Type) []*encoding.Register {
	intRegisterIx := 0
	floatRegisterIx := 0
	var result []*encoding.Register

	var reg *encoding.Register
	for _, arg := range args {
		if arg.Type() == T_Float64 {
			reg = a.floatTargets[floatRegisterIx]
			floatRegisterIx += 1
		} else {
			reg = a.intTargets[intRegisterIx]
			intRegisterIx += 1
		}
		result = append(result, reg)
	}
	return result
}

func (a *ABI_AMDSystemV) ReturnTypeToOperand(arg Type) lib.Operand {
	if arg.Type() == T_Float64 {
		return encoding.Xmm0
	}
	return encoding.Rax
}

// PreserveRegisters returns instructions and clobbered registers
func PreserveRegisters(ctx *IR_Context, argTypes []Type, returnType Type) (lib.Instructions, map[lib.Operand]lib.Operand, []lib.Operand) {
	var clobbered []lib.Operand
	var result []lib.Instruction
	mapping := map[lib.Operand]lib.Operand{}

	// push the return register; TODO: check if in use?
	returnOp := ctx.ABI.ReturnTypeToOperand(returnType)
	push := x86_64.PUSH(returnOp)
	result = append(result, push)
	clobbered = append(clobbered, returnOp)
	ctx.AddInstruction(push)

	allocator := ctx.Allocator.(*X86_64_Allocator)
	// Push registers that are already in use
	regs := ctx.ABI.GetRegistersForArgs(argTypes)
	var inUse bool
	for i, arg := range argTypes {
		reg := regs[i]
		if arg.Type() == T_Float64 {
			inUse = allocator.FloatRegisters[reg.Register]
		} else {
			inUse = allocator.Registers[reg.Register]
		}
		if inUse {
			result = append(result, x86_64.PUSH(reg))
			ctx.AddInstruction(x86_64.PUSH(reg))
			clobbered = append(clobbered, reg)
		}
	}
	// Build the register -> location on the stack mapping
	mappedClobbered := 1 // set to 1, to account for the return op TODO
	for i, arg := range argTypes {
		reg := regs[i]
		if arg.Type() == T_Float64 {
			inUse = allocator.FloatRegisters[reg.Register]
		} else {
			inUse = allocator.Registers[reg.Register]
		}
		if inUse {
			offset := (len(clobbered) - mappedClobbered) * int(arg.Width())
			mapping[reg] = &encoding.DisplacedRegister{Register: encoding.Rsp, Displacement: uint8(offset)}
			mappedClobbered += 1
		}
	}
	return result, mapping, clobbered
}

//goland:noinspection GoSnakeCaseUsage,GoErrorStringFormat
func ABI_Call_Setup(ctx *IR_Context, args []IRExpression, returnType Type) (lib.Instructions, map[lib.Operand]lib.Operand, []lib.Operand, error) {
	argTypes := make([]Type, len(args))
	for i, arg := range args {
		argTypes[i] = arg.ReturnType(ctx)
		if argTypes[i] == nil {
			//goland:noinspection GoErrorStringFormat
			return nil, nil, nil, fmt.Errorf("Unknown type for value: %s", arg)
		}
	}
	result, mapping, clobbered := PreserveRegisters(ctx, argTypes, returnType)
	regs := ctx.ABI.GetRegistersForArgs(argTypes)

	ctx_ := ctx.Copy()
	allocator := ctx_.Allocator.(*X86_64_Allocator)
	for _, reg := range regs {
		if reg.Size == lib.OWORD {
			allocator.FloatRegisters[reg.Register] = true
			allocator.FloatRegistersAllocated += 1
		} else {
			if !allocator.Registers[reg.Register] {
				allocator.Registers[reg.Register] = true
				allocator.RegistersAllocated += 1
			}
		}
	}

	for i, arg := range args {
		// TODO: this should probably move to the "encode" package
		if ctx.Architecture == nil {
			return nil, nil, nil, fmt.Errorf("Missing Architecture in IR_Context")
		}
		instr, err := ctx.Architecture.EncodeExpression(arg, ctx_, regs[i])
		if err != nil {
			return nil, nil, nil, err
		}
		ctx.AddInstruction(instr...)
		result = result.Add(instr)
	}
	for variable, location := range ctx_.VariableMap {
		if newLocation, found := mapping[location]; found {
			ctx_.VariableMap[variable] = newLocation
		}
	}

	return result, mapping, clobbered, nil

}

func RestoreRegisters(ctx *IR_Context, clobbered []lib.Operand) lib.Instructions {
	// Pop in reverse order
	var result []lib.Instruction
	for j := len(clobbered) - 1; j >= 0; j-- {
		reg := clobbered[j]
		result = append(result, x86_64.POP(reg))
		ctx.AddInstruction(x86_64.POP(reg))
	}
	return result
}
