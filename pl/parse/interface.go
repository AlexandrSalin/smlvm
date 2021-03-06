package parse

import (
	"shanhu.io/smlvm/lexing"
	"shanhu.io/smlvm/pl/ast"
)

func parseInterface(p *parser) *ast.Interface {
	if !p.SeeKeyword("interface") {
		panic("expect keyword")
	}

	ret := &ast.Interface{
		Kw:     p.Shift(),
		Name:   p.Expect(Ident),
		Lbrace: p.ExpectOp("{"),
	}

	for !p.SeeOp("}") && !p.See(lexing.EOF) {
		name := p.Expect(Ident)
		if p.InError() {
			return nil
		}

		f := parseFuncSig(p)
		if p.InError() {
			return nil
		}

		ret.Funcs = append(ret.Funcs, &ast.InterfaceFunc{
			Name:    name,
			FuncSig: f,
			Semi:    p.ExpectSemi(),
		})

	}

	ret.Rbrace = p.ExpectOp("}")
	ret.Semi = p.ExpectSemi()
	return ret
}
