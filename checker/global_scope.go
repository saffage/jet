package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

var Global = NewScope(nil)

var primitives = [...]Symbol{
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.Primitives[types.Bool], "bool"),
		node:  nil,
		name:  &ast.Ident{Name: "bool"},
	},
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.Primitives[types.I32], "i32"),
		node:  nil,
		name:  &ast.Ident{Name: "i32"},
	},
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.Primitives[types.I32], "int"),
		node:  nil,
		name:  &ast.Ident{Name: "int"},
	},
}

func (check *Checker) defPrimitives() {
	for _, primitive := range primitives {
		defined := check.module.scope.Define(primitive)
		assert.Ok(defined == nil)
	}
}