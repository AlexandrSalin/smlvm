package tast

import (
	"shanhu.io/smlvm/pl/types"
	"shanhu.io/smlvm/syms"
)

// Func is a function.
type Func struct {
	Sym *syms.Symbol // function symbol

	This     *types.Pointer
	Receiver *syms.Symbol // explicit receiver

	Args      []*syms.Symbol
	NamedRets []*syms.Symbol

	Body []Stmt
}

// IsMethod returns true when the function is a method.
func (f *Func) IsMethod() bool {
	return !(f.This == nil && f.Receiver == nil)
}

// FuncAlias is a function alias.
type FuncAlias struct {
	Sym *syms.Symbol
	Of  *syms.Symbol
}
