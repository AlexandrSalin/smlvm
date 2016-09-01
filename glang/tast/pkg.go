package tast

import (
	"e8vm.io/e8vm/sym8"
)

// Import is an import statement
type Import struct {
	Sym *sym8.Symbol
}

// Pkg is a package of imports, consts, structs, vars and funcs.
type Pkg struct {
	Imports []*sym8.Symbol
	Consts  []*sym8.Symbol
	Structs []*sym8.Symbol

	Vars        []*Define
	FuncAliases []*FuncAlias
	Funcs       []*Func
	Methods     []*Func
}
