package codegen

import (
	"fmt"

	"e8vm.io/e8vm/link8"
)

func copyRef(g *gener, b *Block, dest, src Ref, isArg bool) {
	loadDestAddr := func(r uint32) {
		if !isArg {
			loadAddr(b, r, dest)
		} else {
			loadArgAddr(b, r, dest.(*Var))
		}
	}

	size := dest.Size()
	switch {
	case size != src.Size():
		e := fmt.Errorf("copyRef src(%T)=%d dest(%T)=%d",
			src, src.Size(), dest, size)
		panic(e)
	case size < 0:
		panic("negative size for copyRef")
	case size == 0:
		return
	case canViaReg(dest) && canViaReg(src):
		loadRef(b, _4, src)
		if !isArg {
			saveRef(b, _4, dest, _1)
		} else {
			saveArg(b, _4, dest.(*Var))
		}
	default:
		loadDestAddr(_1)
		loadAddr(b, _2, src)
		loadUint32(b, _3, uint32(size))

		jal := b.inst(asm.jal(0))
		f := g.memCopy
		jal.sym = &linkSym{link8.FillLink, f.pkg, f.name}
	}
}