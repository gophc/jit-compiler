package shared

import (
	"github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	"github.com/bspaans/jit-compiler/lib"
)

type Architecture interface {
	EncodeExpression(expr IRExpression, ctx *IR_Context, target encoding.Operand) ([]lib.Instruction, error)
	EncodeStatement(stmt IR, ctx *IR_Context) ([]lib.Instruction, error)
	EncodeDataSection(stmts []IR, ctx *IR_Context) (*Segments, error)
}

type IR_Context struct {
	Architecture            Architecture
	ABI                     ABI
	Registers               []bool
	RegistersAllocated      uint8
	FloatRegisters          []bool
	FloatRegistersAllocated uint8
	VariableMap             map[string]encoding.Operand
	VariableTypes           map[string]Type
	ReturnOperandStack      []encoding.Operand
	Segments                *Segments
	InstructionPointer      uint
	StackPointer            int
	Commit                  bool // if false turns AddInstruction into a noop

	instructions []lib.Instruction
}

func NewIRContext(arch Architecture) *IR_Context {
	ctx := &IR_Context{
		Architecture:            arch,
		ABI:                     NewABI_AMDSystemV(),
		Registers:               make([]bool, 16),
		RegistersAllocated:      0,
		FloatRegisters:          make([]bool, 16),
		FloatRegistersAllocated: 0,
		VariableMap:             map[string]encoding.Operand{},
		VariableTypes:           map[string]Type{},
		ReturnOperandStack:      []encoding.Operand{&encoding.DisplacedRegister{encoding.Rsp, 8}},
		InstructionPointer:      2,
		StackPointer:            8,
		Commit:                  true,
		instructions:            []lib.Instruction{},
	}
	// Always allocate rsp
	// Should track usage?
	ctx.Registers[4] = true // stack pointer
	ctx.Registers[5] = true // frame pointer
	ctx.RegistersAllocated = 1
	return ctx
}

func (i *IR_Context) PushReturnOperand(op encoding.Operand) {
	i.ReturnOperandStack = append(i.ReturnOperandStack, op)
}
func (i *IR_Context) PeekReturn() encoding.Operand {
	return i.ReturnOperandStack[len(i.ReturnOperandStack)-1]
}

func (i *IR_Context) PopReturn() encoding.Operand {
	op := i.ReturnOperandStack[len(i.ReturnOperandStack)-1]
	i.ReturnOperandStack = i.ReturnOperandStack[:len(i.ReturnOperandStack)-1]
	return op
}

func (i *IR_Context) Copy() *IR_Context {
	regs := make([]bool, 16)
	floatRegs := make([]bool, 16)
	for j := 0; j < 16; j++ {
		regs[j] = i.Registers[j]
		floatRegs[j] = i.FloatRegisters[j]
	}
	variableMap := map[string]encoding.Operand{}
	for arg, reg := range i.VariableMap {
		variableMap[arg] = reg
	}
	variableTypes := map[string]Type{}
	for arg, ty := range i.VariableTypes {
		variableTypes[arg] = ty
	}
	instructions := []lib.Instruction{}
	for _, d := range i.instructions {
		instructions = append(instructions, d)
	}
	returns := []encoding.Operand{}
	for _, d := range i.ReturnOperandStack {
		returns = append(returns, d)
	}
	return &IR_Context{
		Architecture:            i.Architecture,
		ABI:                     i.ABI,
		Registers:               regs,
		RegistersAllocated:      i.RegistersAllocated,
		FloatRegisters:          floatRegs,
		FloatRegistersAllocated: i.FloatRegistersAllocated,
		VariableMap:             variableMap,
		VariableTypes:           variableTypes,
		ReturnOperandStack:      returns,
		Segments:                i.Segments,
		InstructionPointer:      i.InstructionPointer,
		StackPointer:            i.StackPointer,
		Commit:                  i.Commit,
		instructions:            instructions,
	}
}

func (i *IR_Context) AddInstruction(instr ...lib.Instruction) {
	if i.Commit {
		for _, in := range instr {
			i.instructions = append(i.instructions, in)
			length, _ := lib.Instruction_Length(in)
			i.InstructionPointer += uint(length)
		}
	}
}

func (i *IR_Context) AddInstructions(instr []lib.Instruction) {
	for _, inst := range instr {
		i.AddInstruction(inst)
	}
}

func (i *IR_Context) GetInstructions() []lib.Instruction {
	return i.instructions
}

func (i *IR_Context) AllocateRegister(typ Type) *encoding.Register {
	if typ == TFloat64 {
		return encoding.GetFloatingPointRegisterByIndex(i.allocateFloatRegister())
	}
	return encoding.Get64BitRegisterByIndex(i.allocateRegister()).ForOperandWidth(typ.Width())
}

func (i *IR_Context) DeallocateRegister(reg *encoding.Register) {
	if reg.Size == lib.QUADDOUBLE {
		i.deallocateFloatRegister(reg.Register)
		return
	}
	i.deallocateRegister(reg.Register)
}

func (i *IR_Context) allocateRegister() uint8 {
	if i.RegistersAllocated >= 16 {
		panic("Register allocation limit. Needs stack handling")
	}
	for j := 0; j < len(i.Registers); j++ {
		if !i.Registers[j] {
			i.Registers[j] = true
			i.RegistersAllocated += 1
			return uint8(j)
		}
	}
	panic("Register allocation limit reached with incorrect allocation counter. Needs stack handling")
}

func (i *IR_Context) deallocateRegister(reg uint8) {
	i.Registers[reg] = false
	i.RegistersAllocated -= 1
}

func (i *IR_Context) allocateFloatRegister() uint8 {
	if i.FloatRegistersAllocated >= 16 {
		panic("FloatRegister allocation limit. Needs stack handling")
	}
	for j := 0; j < len(i.FloatRegisters); j++ {
		if !i.FloatRegisters[j] {
			i.FloatRegisters[j] = true
			i.FloatRegistersAllocated += 1
			return uint8(j)
		}
	}
	panic("FloatRegister allocation limit reached with incorrect allocation counter. Needs stack handling")
}

func (i *IR_Context) deallocateFloatRegister(reg uint8) {
	i.FloatRegisters[reg] = false
	i.FloatRegistersAllocated -= 1
}
