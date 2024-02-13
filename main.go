package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/bspaans/jit-compiler/ir"
	"github.com/bspaans/jit-compiler/ir/encoding/x86_64"
	"github.com/bspaans/jit-compiler/ir/shared"
)

//goland:noinspection GoBoolExpressions
func REPL() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		statements, err := ir.ParseIR(ir.Stdlib + text)
		if err != nil {
			fmt.Println("Parse error: ", err.Error())
			continue

		}
		debug := true
		statements = statements.SSA_Transform(shared.NewSSA_Context())
		instr, err := ir.Compile(&x86_64.X86_64{}, x86_64.NewABI_AMDSystemV(), []shared.IR{statements}, debug)
		if err != nil {
			fmt.Println("Compile error: ", err.Error())
			continue

		}
		fmt.Println(instr.Execute(debug))
	}
}

//goland:noinspection GoBoolExpressions
func CompileFiles() {
	source := ""
	for _, file := range os.Args[1:] {
		text, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		source += string(text) + "\n"
	}
	statements, err := ir.ParseIR(ir.Stdlib + source)
	if err != nil {
		panic(err)
	}

	debug := true
	statements = statements.SSA_Transform(shared.NewSSA_Context())
	if err := ir.CompileToBinary(&x86_64.X86_64{}, x86_64.NewABI_AMDSystemV(), []shared.IR{statements}, debug, "test.bin"); err != nil {
		panic(err)

	}
}

func main() {
	if len(os.Args) == 1 {
		REPL()
	} else {
		CompileFiles()
	}
}
