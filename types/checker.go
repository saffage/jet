package types

import (
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

type checker struct {
	module   *Module
	env      *Env
	cfg      *config.Config
	problems []error
}

func (check *checker) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.LetDecl:
		check.resolveLetDecl(node)

	case *ast.TypeDecl:
		check.resolveTypeDecl(node)

	default:
		panic(&errorIllFormedAst{node})
	}

	return nil
}

// Checks 'expr' and returns it's type. If dest if provided, it also
// will check if expr is of type 'dest', otherwise returns nil.
//
// Also, the value of the expression will also be evaluated (if it's possible)
// and stored in the [TypeInfo.Values] field.
//
// If error was occurred, result is undefined.
func (check *checker) typeOf(expr ast.Node, dest ...Type) (Type, error) {
	if len(dest) > 1 {
		panic("invalid arguments count, expected dest len < 2")
	}

	var expected Type

	if len(dest) == 1 && dest[0] != nil {
		expected = dest[0]
	}

	if value, err := check.valueOf(expr, dest...); err != nil {
		return nil, err
	} else if value != nil {
		return value.T, nil
	}

	t, err := check.typeOfInternal(expr)

	if err != nil {
		return nil, err
	}

	if t == nil {
		panic(internalErrorf(expr, "unresolved type of the expression"))
	}

	if expected != nil && !check.convertible(&Value{t, nil}, expected) {
		report.Debug("%T and %T", t, dest[0])
		err = &errorTypeMismatch{expr, nil, t, expected}
	}

	check.setType(expr, t)
	return t, err
}

func (check *checker) valueOf(expr ast.Node, dest ...Type) (*Value, error) {
	if len(dest) > 1 {
		panic("invalid arguments count, expected dest len < 2")
	}

	if t, ok := check.module.Values[expr]; ok {
		return t, nil
	}

	value, err := check.valueOfInternal(expr)

	if err == nil && value != nil {
		if len(dest) > 0 && dest[0] != nil && !check.convertible(value, dest[0]) {
			return nil, &errorValueCannotBeStoredAsX{expr, value.T, dest[0]}
		}
		check.setValue(expr, value)
		return value, nil
	}

	return nil, err
}

func (check *checker) convertible(value *Value, expected Type) (x bool) {
	report.DebugX("convertible", "`%s` to `%s`", value.T, expected)

	if value.T == expected || value.T.Equal(expected) {
		report.DebugX("convertible", "true")
		return true
	}

	typed := IntoTyped(value.T, expected)
	b := typed != nil && typed.Equal(expected)
	report.DebugX("convertible", "typed `%s`: %v", typed, b)
	return b
}

// Used for better readability like:
//
//	defer check.setEnv(check.env)
//	check.setEnv(someEnv)
func (check *checker) setEnv(scope *Env) {
	assert(scope != nil)
	check.env = scope
}

func (check *checker) setType(expr ast.Node, t Type) {
	assert(expr != nil)
	assert(t != nil)

	if prev := check.module.Values[expr]; prev != nil {
		// WARNING is it safe?
		check.module.Values[expr] = &Value{t, prev.V}
	} else {
		check.module.Values[expr] = &Value{t, nil}
	}
}

func (check *checker) setValue(expr ast.Node, value *Value) {
	assert(expr != nil)
	assert(value != nil)
	assert(value.T != nil)

	check.module.Values[expr] = value
}

func (check *checker) newDef(ident ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)

	var symStr string

	if debugPrinter, _ := sym.(debugPrinter); debugPrinter != nil {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}

	report.Debug("def %s `%s`", color.HiBlueString(symStr), ident)
	check.module.Defs.Put(ident, sym)

	// check.setType(ident, sym.Type())

	// switch sym.(type) {
	// case *StructSymbol, *EnumSymbol:
	// 	check.module.TypeSymbols[SkipTypeDesc(sym.Type())] = sym
	// }
}

func (check *checker) newUse(ident ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)

	_, isDef := check.module.Defs.Get(ident)
	assert(!isDef)

	var symStr string

	if debugPrinter, ok := sym.(debugPrinter); ok {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}

	report.Debug(
		"use of %s `%s` in `%s` at %s",
		color.HiBlueString(symStr),
		sym.Ident(),
		ident,
		ident.Pos(),
	)
	check.module.Uses[ident] = sym
}
