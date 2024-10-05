package types

import (
	"fmt"
	"strings"
)

type debugPrinter interface {
	debug() string
}

func (sym *Binding) debug() string {
	result := symbolTypeNoQualifier(sym)
	mods := make([]string, 0, 5)
	if sym.isParam {
		mods = append(mods, "param")
	}
	if sym.isField {
		mods = append(mods, "field")
	}
	if sym.isVariant {
		mods = append(mods, "variant")
	}
	if sym.isGlobal {
		mods = append(mods, "global")
	}
	if sym.isExtern {
		mods = append(mods, "extern")
	}
	if len(mods) != 0 {
		result += "(" + strings.Join(mods, ", ") + ")"
	}
	return result
}

func symbolTypeNoQualifier(sym Symbol) string {
	return strings.TrimPrefix(
		strings.TrimPrefix(fmt.Sprintf("%T", sym), "*"),
		"types.",
	)
}
