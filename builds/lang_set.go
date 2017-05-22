package builds

import (
	"fmt"
	"strings"
)

// LangSet is a utility that picks a language based on a keyword.
type LangSet struct {
	defaultLang *Lang
	langs       map[string]*Lang
}

// NewLangPicker creates a new picker with a default language.
func NewLangPicker(def *Lang) *LangSet {
	if def == nil {
		panic("default language must not be nil")
	}

	return &LangSet{
		defaultLang: def,
		langs:       make(map[string]*Lang),
	}
}

// AddLang adds a language that has a certain keyword.
// When key is empty, it replaces the default language.
func (pick *LangSet) AddLang(key string, lang *Lang) {
	if lang == nil {
		panic("language must not be nil")
	}

	if key == "" {
		pick.defaultLang = lang
	}
	pick.langs[key] = lang
}

// Lang picks the language for a particular path.
func (pick *LangSet) Lang(path string) *Lang {
	if !IsPkgPath(path) {
		panic(fmt.Errorf("%q is not a package path", path))
	}

	pkgs := strings.Split(path, "/")
	for _, pkg := range pkgs {
		if ret, ok := pick.langs[pkg]; ok {
			return ret
		}
	}

	return pick.defaultLang
}