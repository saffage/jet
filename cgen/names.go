package cgen

import (
	"io"
	"strings"

	"github.com/saffage/jet/checker"
)

var names = map[checker.Symbol]string{}

func (gen *generator) name(sym checker.Symbol) string {
	if name, ok := names[sym]; ok {
		return name
	}
	buf := strings.Builder{}

	switch sym := sym.(type) {
	case *checker.Var:
		if sym.IsGlobal() {
			buf.WriteString("g_")
		}
		if sym.IsParam() {
			buf.WriteString("p_")
		}

	case *checker.Func:
		if sym.IsExtern() {
			return sym.Name()
		}
	}

	gen.nameInternal(&buf, sym.Owner())
	buf.WriteString(sym.Name())
	names[sym] = buf.String()
	return buf.String()
}

func (gen *generator) nameInternal(w io.StringWriter, scope *checker.Scope) {
	for scope != nil && scope != checker.Global {
		scopeName := scope.Name()
		spaceIndex := strings.Index(scopeName, " ")

		if spaceIndex == -1 {
			spaceIndex = len(scopeName)
		}

		switch scopeName[:spaceIndex] {
		case "module":
			defer w.WriteString(scopeName[spaceIndex+1:] + "__")

		case "func":
			defer w.WriteString(scopeName[spaceIndex+1:] + "__")

		case "enum":
			defer w.WriteString(scopeName[spaceIndex+1:] + "__")

		case "block":
			// There must be a block ID I think.

		case "struct":
		case "global":
		default:
			// Do nothing.
		}

		scope = scope.Parent()
	}
}
