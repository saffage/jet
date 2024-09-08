package checker

import (
	"fmt"
	"strings"
)

type debugPrinter interface {
	debug() string
}

func (sym *Binding) debug() string {
	result := symbolTypeNoQualifier(sym)
	mods := make([]string, 0, 2)
	if sym.isParam {
		mods = append(mods, "param")
	}
	if sym.isField {
		mods = append(mods, "field")
	}
	if len(mods) != 0 {
		result += "(" + strings.Join(mods, ", ") + ")"
	}
	return result
}

func symbolTypeNoQualifier(sym Symbol) string {
	return strings.TrimPrefix(
		strings.TrimPrefix(fmt.Sprintf("%T", sym), "*"),
		"checker.",
	)
}
