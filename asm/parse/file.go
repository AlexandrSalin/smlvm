// Package parse parses an assembly program into an AST.
package parse

import (
	"io"

	"e8vm.io/e8vm/asm/ast"
	"e8vm.io/e8vm/lexing"
)

func parseFile(p *parser) *ast.File {
	ret := new(ast.File)

	if p.SeeKeyword("import") {
		if imp := parseImports(p); imp != nil {
			ret.Imports = imp
		}
	}

	for !p.See(lexing.EOF) {
		if p.SeeKeyword("func") {
			if f := parseFunc(p); f != nil {
				ret.Decls = append(ret.Decls, f)
			}
		} else if p.SeeKeyword("var") {
			if v := parseVar(p); v != nil {
				ret.Decls = append(ret.Decls, v)
			}
		} else if p.SeeKeyword("const") {
			// TODO(h8liu): support const
			p.ErrorfHere("const support not implemented yet")
			p.skipErrStmt()
		} else {
			p.ErrorfHere("expect top-declaration: func, var or const")
			return nil
		}

		p.BailOut()
	}

	return ret
}

// Result is a file parsing result
type Result struct {
	File   *ast.File
	Tokens []*lexing.Token
}

// FileResult returns a parsing result.
func FileResult(f string, rc io.ReadCloser) (*Result, []*lexing.Error) {
	p, rec := newParser(f, rc)
	parsed := parseFile(p)
	e := rc.Close()

	if e != nil {
		return nil, lexing.SingleErr(e)
	}
	if es := p.Errs(); es != nil {
		return nil, es
	}

	res := &Result{
		File:   parsed,
		Tokens: rec.Tokens(),
	}
	return res, nil
}

// File function parses a file into an AST.
func File(f string, rc io.ReadCloser) (*ast.File, []*lexing.Error) {
	res, es := FileResult(f, rc)
	if es != nil {
		return nil, es
	}
	return res.File, es
}
