package asm8

import (
	"bytes"
	"strconv"

	"e8vm.io/e8vm/asm8/parse"
	"e8vm.io/e8vm/lexing"
)

func parseDataHex(p lexing.Logger, args []*lexing.Token) ([]byte, uint32) {
	if !checkTypeAll(p, args, parse.Operand) {
		return nil, 0
	}

	buf := new(bytes.Buffer)
	for _, arg := range args {
		b, e := strconv.ParseUint(arg.Lit, 16, 8)
		if e != nil {
			p.Errorf(arg.Pos, "%s", e)
			return nil, 0
		}

		buf.WriteByte(byte(b))
	}

	return buf.Bytes(), 0
}
