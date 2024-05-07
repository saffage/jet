package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

type Checker struct {
	*TypeInfo
	module         *Module
	scope          *Scope
	builtIns       []*BuiltIn
	errors         []error
	isErrorHandled bool
}

type TypeInfo struct {
	Data  map[ast.Node]TypedValue
	Types map[ast.Node]TypedValue
	Defs  map[*ast.Ident]Symbol
}

func Check(node *ast.ModuleDecl) (*TypeInfo, []error) {
	module := NewModule(Global, node)
	check := &Checker{
		TypeInfo: &TypeInfo{
			Data:  make(map[ast.Node]TypedValue),
			Types: make(map[ast.Node]TypedValue),
			Defs:  make(map[*ast.Ident]Symbol),
		},
		module:         module,
		scope:          module.scope,
		errors:         make([]error, 0),
		isErrorHandled: true,
	}

	check.defBuiltIns()
	check.defPrimitives()

	{
		nodes := []ast.Node(nil)

		switch body := node.Body.(type) {
		case *ast.List:
			nodes = body.Nodes

		case *ast.CurlyList:
			nodes = body.List.Nodes

		default:
			panic("ill-formed AST")
		}

		for _, node := range nodes {
			ast.WalkTopDown(check.visit, node)
		}

		module.completed = true
	}

	return check.TypeInfo, check.errors
}

// Type checks 'expr' and returns its type.
// If error was occured, result is undefined
func (check *Checker) typeOf(expr ast.Node) types.Type {
	if t, ok := check.Types[expr]; ok {
		return t.Type
	}

	if v := check.valueOfInternal(expr); v != nil {
		check.setValue(expr, *v)
		return v.Type
	}

	if t := check.typeOfInternal(expr); t != nil {
		check.setType(expr, t)
		return t
	}

	return nil
}

func (check *Checker) valueOf(expr ast.Node) *TypedValue {
	if t, ok := check.Types[expr]; ok {
		return &t
	}

	if value := check.valueOfInternal(expr); value != nil {
		check.setValue(expr, *value)
		return value
	}

	return nil
}

func (check *Checker) setScope(scope *Scope) {
	check.scope = scope
}

func (check *Checker) setType(expr ast.Node, t types.Type) {
	assert.Ok(expr != nil)
	assert.Ok(t != nil)

	if check.Types != nil {
		check.Types[expr] = TypedValue{t, nil}
	}
}

func (check *Checker) setValue(expr ast.Node, value TypedValue) {
	assert.Ok(expr != nil)
	assert.Ok(value.Type != nil)

	if check.Types != nil {
		check.Types[expr] = value
	}
}

func (check *Checker) newDef(ident *ast.Ident, sym Symbol) {
	assert.Ok(ident != nil)

	if check.Defs != nil {
		check.Defs[ident] = sym
	}
}
