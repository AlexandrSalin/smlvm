package asm8

import (
	"e8vm.io/e8vm/asm8/ast"
	"e8vm.io/e8vm/lexing"
)

type varDecl struct {
	*ast.Var

	stmts []*varStmt
}

func resolveVar(log lexing.Logger, v *ast.Var) *varDecl {
	ret := new(varDecl)

	ret.Var = v

	for _, stmt := range v.Stmts {
		r := resolveVarStmt(log, stmt)
		ret.stmts = append(ret.stmts, r)
	}

	return ret
}
