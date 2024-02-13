package main_test

import (
	"fmt"
	"github.com/bspaans/jit-compiler/ir"
	"github.com/bspaans/jit-compiler/ir/encoding/x86_64"
	"github.com/bspaans/jit-compiler/ir/shared"
	"testing"
)

//goland:noinspection GoBoolExpressions
func TestCompile(t *testing.T) {
	code := `prev = 1; current = 1;
while current < 13 {
  tmp = current
  current = current + prev
  prev = tmp
}
return current
`

	debug := true
	statements := ir.MustParseIR(code)
	machineCode, err := ir.Compile(&x86_64.X86_64{},
		x86_64.NewABI_AMDSystemV(),
		[]shared.IR{statements},
		debug)
	if err != nil {
		panic(err)
	}
	fmt.Println(machineCode.Execute(debug))
}
