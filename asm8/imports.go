package asm8

import (
	"io"

	"e8vm.io/e8vm/asm8/parse"
	"e8vm.io/e8vm/build8"
	"e8vm.io/e8vm/lexing"
)

func listImport(
	f string, rc io.ReadCloser, imp build8.Importer,
) []*lexing.Error {
	astFile, es := parse.File(f, rc)
	if es != nil {
		return es
	}

	if astFile.Imports == nil {
		return nil
	}

	log := lexing.NewErrorList()
	impDecl := resolveImportDecl(log, astFile.Imports)
	if es := log.Errs(); es != nil {
		return es
	}

	for as, stmt := range impDecl.stmts {
		imp.Import(as, stmt.path, stmt.Path.Pos)
	}

	return nil
}
