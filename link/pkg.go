package link

import (
	"fmt"
	"io"
)

// Pkg is the compiling object of a package. It is the linking
// unit for programs.
type Pkg struct {
	path string

	symbols map[string]*Symbol // all the symbol objects

	funcs map[string]*Func
	vars  map[string]*Var

	funcMetas map[string]*FuncMeta
}

// NewPkg creates a new package for path p.
func NewPkg(p string) *Pkg {
	ret := &Pkg{
		path:      p,
		symbols:   make(map[string]*Symbol),
		funcs:     make(map[string]*Func),
		vars:      make(map[string]*Var),
		funcMetas: make(map[string]*FuncMeta),
	}

	return ret
}

// Path returns the package's path string.
func (p *Pkg) Path() string { return p.path }

// Declare declares a symbol and assigns a symbol index.
// If s.Name is empty string, then the symbol is anonymous.
func (p *Pkg) declare(s *Symbol) {
	if s.Name == "" {
		panic("empty symbol name")
	}

	_, found := p.symbols[s.Name]
	if found {
		panic("symbol redeclare")
	}
	p.symbols[s.Name] = s
}

// DeclareFunc declares a function (code block).
func (p *Pkg) DeclareFunc(name string) {
	if name == "" {
		panic("name empty")
	}
	p.declare(&Symbol{name, SymFunc})
}

// DeclareVar declares a variable (data block)
func (p *Pkg) DeclareVar(name string) {
	p.declare(&Symbol{name, SymVar})
}

// SymbolByName returns the symbol with the particular name.
func (p *Pkg) SymbolByName(name string) *Symbol {
	return p.symbols[name]
}

// HasFunc checks if the package has a function of a particular name.
func (p *Pkg) HasFunc(name string) bool {
	sym := p.SymbolByName(name)
	if sym == nil || sym.Type != SymFunc {
		return false
	}
	return true
}

// DefineFunc instantiates a function object.
func (p *Pkg) DefineFunc(name string, f *Func) {
	sym := p.SymbolByName(name)
	if sym.Type != SymFunc {
		panic("not a function")
	}
	p.funcs[name] = f
}

// DefineVar instantiates a variable object.
func (p *Pkg) DefineVar(name string, v *Var) {
	sym := p.SymbolByName(name)
	if sym == nil {
		panic(fmt.Errorf("symbol %q not found", name))
	}
	if sym.Type != SymVar {
		panic("not a var")
	}
	p.vars[name] = v
}

// Func returns the function of index.
func (p *Pkg) Func(name string) *Func {
	ret, found := p.funcs[name]
	if !found {
		panic("not defined")
	}
	return ret
}

// Var returns the variable of index.
func (p *Pkg) Var(name string) *Var {
	ret, found := p.vars[name]
	if !found {
		panic("not defined")
	}
	return ret
}

// PrintSymbols prints all symbols out to a writer.
func (p *Pkg) PrintSymbols(out io.Writer) {
	for index, sym := range p.symbols {
		fmt.Fprintf(out, "%d: %s %s\n",
			index, symStr(sym.Type), sym.Name,
		)
	}
}

func pkgVar(pkgs map[string]*Pkg, ps *PkgSym) *Var {
	return pkgs[ps.Pkg].Var(ps.Sym)
}

func pkgFunc(pkgs map[string]*Pkg, ps *PkgSym) *Func {
	return pkgs[ps.Pkg].Func(ps.Sym)
}

func pkgSym(pkgs map[string]*Pkg, ps *PkgSym) *Symbol {
	return pkgs[ps.Pkg].SymbolByName(ps.Sym)
}
