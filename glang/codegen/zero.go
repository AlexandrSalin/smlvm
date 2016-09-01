package codegen

import (
	"fmt"

	"e8vm.io/e8vm/link8"
)

func zeroAddr(g *gener, b *Block, reg uint32, size int32, regSizeAlign bool) {
	switch {
	case size < 4:
		for i := int32(0); i < size; i++ {
			b.inst(asm.sb(_r0, _r1, i))
		}
	case size == 4 && regSizeAlign:
		b.inst(asm.sw(_r0, _r1, 0))
	case size == 8 && regSizeAlign:
		b.inst(asm.sw(_r0, _r1, 0))
		b.inst(asm.sw(_r0, _r1, 4))
	default:
		loadUint32(b, _r2, uint32(size))
		jal := b.inst(asm.jal(0))
		f := g.memClear
		jal.sym = &linkSym{link8.FillLink, f.pkg, f.name}
	}
}

func zeroRef(g *gener, b *Block, r Ref) {
	switch r := r.(type) {
	case *Var:
		if r.size < 0 {
			panic("invalid varRef size")
		}

		switch r.size {
		case 0: // do nothing
		case 1, regSize:
			saveVar(b, 0, r)
		default:
			loadAddr(b, _r1, r)
			loadUint32(b, _r2, uint32(r.size))

			jal := b.inst(asm.jal(0))
			f := g.memClear
			jal.sym = &linkSym{link8.FillLink, f.pkg, f.name}
		}
	case *AddrRef:
		if r.size == 0 {
			return
		}
		loadAddr(b, _r1, r)
		zeroAddr(g, b, _r1, r.size, r.regSizeAlign)
	case *HeapSym:
		if r.size == 0 {
			return
		}
		loadAddr(b, _r1, r)
		zeroAddr(g, b, _r1, r.size, true)
	case *number:
		panic("number are read only")
	default:
		panic(fmt.Errorf("not implemented: %T", r))
	}
}

// CanBeZero checks if a reference could be zero value.
func CanBeZero(r Ref) bool {
	switch r := r.(type) {
	case *number:
		return r.v == 0
	case *FuncSym:
		return false
	case *Func:
		return false
	}
	return true
}
