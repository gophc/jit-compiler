package statements

import (
	"fmt"

	. "github.com/bspaans/jit-compiler/ir/shared"
)

//goland:noinspection GoSnakeCaseUsage
type IR_Return struct {
	*BaseIR
	Expr IRExpression
}

//goland:noinspection GoSnakeCaseUsage
func NewIR_Return(expr IRExpression) *IR_Return {
	return &IR_Return{
		BaseIR: NewBaseIR(Return),
		Expr:   expr,
	}
}

func (i *IR_Return) String() string {
	return fmt.Sprintf("return %s", i.Expr.String())
}

//goland:noinspection GoSnakeCaseUsage
func (i *IR_Return) SSA_Transform(ctx *SSA_Context) IR {
	rewrites, expr := i.Expr.SSA_Transform(ctx)
	ir := SSA_Rewrites_to_IR(rewrites)
	if ir == nil {
		return i
	}
	return NewIR_AndThen(ir, NewIR_Return(expr))
}
