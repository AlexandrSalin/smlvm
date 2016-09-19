package pl

import (
	"shanhu.io/smlvm/pl/codegen"
	"shanhu.io/smlvm/pl/tast"
	"shanhu.io/smlvm/pl/types"
)

func arrayElementSize(t types.T) int32 {
	ret := t.Size()
	if t.RegSizeAlign() {
		return types.RegSizeAlignUp(ret)
	}
	return ret
}

func checkInRange(b *builder, index, n codegen.Ref, op string) {
	inRange := b.newCond()
	b.b.Arith(inRange, index, op, n)

	outOfRange := b.f.NewBlock(b.b)
	after := b.f.NewBlock(outOfRange)
	b.b.JumpIf(inRange, after)

	b.b = outOfRange
	callPanic(b, "index out of range")

	b.b = after
}

func newSlice(b *builder, t types.T, addr, size codegen.Ref) *ref {
	ret := b.newTemp(&types.Slice{T: t})
	retAddr := b.newPtr()
	b.b.Arith(retAddr, nil, "&", ret.IR())
	b.b.Assign(codegen.NewAddrRef(retAddr, 4, 0, false, true), addr)
	b.b.Assign(codegen.NewAddrRef(retAddr, 4, 4, false, true), size)
	return ret
}

func buildIndexExpr(b *builder, expr *tast.IndexExpr) *ref {
	array := b.buildExpr(expr.Array)

	if expr.HasColon {
		return buildSlicing(b, expr, array)
	}
	return buildArrayGet(b, expr, array)
}

func loadArray(b *builder, array *ref) (addr, n codegen.Ref, et types.T) {
	base := b.newPtr()
	t := array.Type()
	switch t := t.(type) {
	case *types.Array:
		b.b.Arith(base, nil, "&", array.IR())
		return base, codegen.Snum(t.N), t.T
	case *types.Slice:
		b.b.Arith(base, nil, "&", array.IR())
		addr = codegen.NewAddrRef(base, 4, 0, false, true)
		n = codegen.NewAddrRef(base, 4, 4, false, true)
		return addr, n, t.T
	}
	panic("bug")
}

func checkArrayIndex(b *builder, index *ref) codegen.Ref {
	t := index.Type()
	if types.IsSigned(t) {
		neg := b.newCond()
		b.b.Arith(neg, nil, "<0", index.IR())
		negPanic := b.f.NewBlock(b.b)
		after := b.f.NewBlock(negPanic)
		b.b.JumpIfNot(neg, after)

		b.b = negPanic
		callPanic(b, "index is negative")

		b.b = after
	}
	return index.IR()
}

func buildArrayIndex(b *builder, expr tast.Expr) codegen.Ref {
	index := b.buildExpr(expr)
	return checkArrayIndex(b, index)
}

func buildSlicing(b *builder, expr *tast.IndexExpr, array *ref) *ref {
	baseAddr, n, et := loadArray(b, array)

	var addr, indexStart, offset codegen.Ref
	if expr.Index == nil {
		indexStart = codegen.Num(0)
		addr = baseAddr
	} else {
		indexStart = buildArrayIndex(b, expr.Index)
		checkInRange(b, indexStart, n, "u<=")

		offset = b.newPtr()
		b.b.Arith(offset, indexStart, "*", codegen.Snum(arrayElementSize(et)))
		addr = b.newPtr()
		b.b.Arith(addr, baseAddr, "+", offset)
	}

	var indexEnd codegen.Ref
	if expr.IndexEnd == nil {
		indexEnd = n
	} else {
		indexEnd = buildArrayIndex(b, expr.IndexEnd)
		checkInRange(b, indexEnd, n, "u<=")
		checkInRange(b, indexStart, indexEnd, "u<=")
	}

	size := b.newPtr()
	b.b.Arith(size, indexEnd, "-", indexStart)
	return newSlice(b, et, addr, size)
}

func buildArrayGet(b *builder, expr *tast.IndexExpr, array *ref) *ref {
	index := buildArrayIndex(b, expr.Index)
	base, n, et := loadArray(b, array)
	checkInRange(b, index, n, "u<")

	addr := b.newPtr()
	b.b.Arith(addr, index, "*", codegen.Snum(arrayElementSize(et)))
	b.b.Arith(addr, base, "+", addr)
	size := et.Size()

	retIR := codegen.NewAddrRef(
		addr,             // base address
		size,             // size
		0,                // dynamic offset; precalculated
		types.IsByte(et), // isByte
		true,             // isAlign
	)
	return newAddressableRef(et, retIR)
}
