package asm8

import (
	"e8vm.io/e8vm/asm8/parse"
	"e8vm.io/e8vm/lexing"
)

var regNameMap = map[string]uint32{
	"r0": 0,
	"r1": 1,
	"r2": 2,
	"r3": 3,
	"r4": 4,
	"r5": 5,
	"r6": 6,
	"r7": 7,

	"sp":  5,
	"ret": 6,
	"pc":  7,
}

func resolveReg(p lexing.Logger, op *lexing.Token) uint32 {
	if op.Type != parse.Operand {
		panic("not an operand")
	}

	ret, found := regNameMap[op.Lit]
	if !found {
		p.Errorf(op.Pos, "invalid register name %q", op.Lit)
		return 0
	}
	return ret
}
