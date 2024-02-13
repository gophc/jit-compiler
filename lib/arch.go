package lib

// newArchContext returns a new archContext which is architecture-specific type to be embedded in callEngine.
// This must be initialized in init() function in architecture-specific arch_*.go file which is guarded by build tag.
var newArchContext func() archContext

type callEngine struct {
	i64 int64
}

type ModuleInstance struct {
	i64 int64
}

// nativecall is used by callEngine.execWasmFunction and the entrypoint to enter the compiled native code.
// codeSegment is the pointer to the initial instruction of the compiled native code.
//
// Note: this is implemented in per-arch Go assembler file. For example, arch_amd64.s implements this for amd64.
func nativecall(codeSegment uintptr, ce *callEngine) int
