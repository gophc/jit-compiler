#include "funcdata.h"
#include "textflag.h"

// nativecall(codeSegment, ce)
TEXT ·nativecall(SB), NOSPLIT|NOFRAME, $0-24
	MOVQ ce+8(FP), R13                     // Load the address of *callEngine. into amd64ReservedRegisterForCallEngine.
	MOVQ codeSegment+0(FP), AX             // Load the address of native code.
	JMP  AX                                // Jump to native code.
