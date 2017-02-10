package codegen

import (
	"fmt"

	"shanhu.io/smlvm/link"
)

func loadRetAddr(b *Block, v *Var) {
	if v.size != regSize {
		panic("ret must be regsize")
	}
	loadArg(b, _pc, v)
}

func saveRetAddr(b *Block, v *Var) {
	if v.size != regSize {
		panic("ret must be regsize")
	}
	saveArg(b, _ret, v)
}

func saveArg(b *Block, reg uint32, v *Var) {
	switch v.size {
	case 0:
	case 1:
		b.inst(asm.sb(reg, _sp, -v.Offset))
	case regSize:
		b.inst(asm.sw(reg, _sp, -v.Offset))
	default:
		panic("invalid size to save from a register")
	}
}

func loadArg(b *Block, reg uint32, v *Var) {
	switch v.size {
	case 0:
	case 1:
		if !v.U8 {
			b.inst(asm.lb(reg, _sp, -v.Offset))
		} else {
			b.inst(asm.lbu(reg, _sp, -v.Offset))
		}
	case regSize:
		b.inst(asm.lw(reg, _sp, -v.Offset))
	default:
		panic("invalid size to save from a register")
	}
}

func saveVar(b *Block, reg uint32, v *Var) {
	switch v.size {
	case 0:
	case 1:
		b.inst(asm.sb(reg, _sp, *b.frameSize-v.Offset))
	case regSize:
		b.inst(asm.sw(reg, _sp, *b.frameSize-v.Offset))
	default:
		panic("invalid size to save from a register")
	}
}

func loadVar(b *Block, reg uint32, v *Var) {
	switch v.size {
	case 0: // do nothing
	case 1:
		if !v.U8 {
			b.inst(asm.lb(reg, _sp, *b.frameSize-v.Offset))
		} else {
			b.inst(asm.lbu(reg, _sp, *b.frameSize-v.Offset))
		}
	default:
		b.inst(asm.lw(reg, _sp, *b.frameSize-v.Offset))
	}
}

func saveRef(b *Block, reg uint32, r Ref, tmpReg uint32) {
	if reg == tmpReg {
		panic("cannot use the same reg")
	}

	switch r := r.(type) {
	case *Var:
		saveVar(b, reg, r)
	case *AddrRef:
		if r.size == 0 {
			return
		}

		loadRef(b, tmpReg, r.base)
		if r.size == 1 {
			b.inst(asm.sb(reg, tmpReg, r.offset))
		} else if r.size == regSize && r.regSizeAlign {
			b.inst(asm.sw(reg, tmpReg, r.offset))
		} else {
			panic("invalid addrRef to save via register")
		}
	case *HeapSym:
		if r.size == 0 {
			return
		}
		loadSym(b, tmpReg, r.pkg, r.name)
		if r.size == 1 {
			b.inst(asm.sb(reg, tmpReg, 0))
		} else if r.size == regSize {
			b.inst(asm.sw(reg, tmpReg, 0))
		} else {
			panic("invalid heapSym to save via register")
		}
	case *FuncPtr:
		saveRef(b, reg, r.Ref, tmpReg)
	case *number, *byt:
		panic("constant references are read only")
	default:
		panic("not implemented")
	}
}

func loadSym(b *Block, reg uint32, pkg, sym string) {
	high := b.inst(asm.addui(reg, _r0, 0))
	high.sym = &linkSym{link.FillHigh, pkg, sym}
	low := b.inst(asm.ori(reg, reg, 0))
	low.sym = &linkSym{link.FillLow, pkg, sym}
}

func loadUint32(b *Block, reg uint32, v uint32) {
	high := v >> 16
	if high != 0 {
		b.inst(asm.addui(reg, _r0, high))
		b.inst(asm.ori(reg, reg, v))
	} else {
		b.inst(asm.ori(reg, _r0, v))
	}
}

func loadRef(b *Block, reg uint32, r Ref) {
	switch r := r.(type) {
	case *Var:
		loadVar(b, reg, r)
	case *number:
		loadUint32(b, reg, r.v)
	case *byt:
		if r.u8 {
			b.inst(asm.ori(reg, _r0, uint32(r.v)))
		} else {
			bt := int8(r.v)
			if bt >= 0 {
				b.inst(asm.ori(reg, _r0, uint32(bt)))
			} else {
				b.inst(asm.addui(reg, _r0, 0xffff))
				b.inst(asm.ori(reg, reg, uint32(int32(bt))&0xffff))
			}
		}
	case *Func:
		loadSym(b, reg, r.pkg, r.name)
	case *FuncSym:
		loadSym(b, reg, r.pkg, r.name)
	case *FuncPtr:
		loadRef(b, reg, r.Ref)
	case *AddrRef:
		if r.size == 0 {
			return
		}

		loadRef(b, reg, r.base)
		if r.size == 1 {
			if r.u8 {
				b.inst(asm.lbu(reg, reg, r.offset))
			} else {
				b.inst(asm.lb(reg, reg, r.offset))
			}
		} else if r.size == regSize && r.regSizeAlign {
			b.inst(asm.lw(reg, reg, r.offset))
		} else if !r.regSizeAlign {
			panic("not reg size aligned addrRef")
		} else { // r.size != regSize
			panic("addrRef not reg size to load via register")
		}
	case *HeapSym:
		if r.size == 0 {
			return
		}
		loadSym(b, reg, r.pkg, r.name)
		if r.size == 1 {
			if r.u8 {
				b.inst(asm.lb(reg, reg, 0))
			} else {
				b.inst(asm.lbu(reg, reg, 0))
			}
		} else if r.size == regSize {
			b.inst(asm.lw(reg, reg, 0))
		} else {
			panic("invalid heapSym to load via register")
		}
	default:
		panic(fmt.Errorf("not implemented: %T", r))
	}
}

func canViaReg(r Ref) bool {
	switch r := r.(type) {
	case *Var:
		return r.size <= 1 || r.size == regSize
	case *number:
		return true
	case *byt:
		return true
	case *Func:
		return true
	case *FuncSym:
		return true
	case *AddrRef:
		return r.size <= 1 || (r.size == regSize && r.regSizeAlign)
	case *HeapSym:
		return r.size <= 1 || r.size == regSize
	}
	return false
}
