package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

var Global = NewScope(nil)

var primitives = [...]Symbol{
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.Bool, "bool"),
		node:  nil,
		name:  &ast.Ident{Name: "bool"},
	},
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.I32, "i32"),
		node:  nil,
		name:  &ast.Ident{Name: "i32"},
	},
	&TypeAlias{
		owner: nil,
		t:     types.NewAlias(types.I32, "int"),
		node:  nil,
		name:  &ast.Ident{Name: "int"},
	},
	NewConst(
		nil,
		&TypedValue{
			Type:  types.UntypedBool,
			Value: constant.NewBool(true),
		},
		&ast.Ident{Name: "true"},
	),
	NewConst(
		nil,
		&TypedValue{
			Type:  types.UntypedBool,
			Value: constant.NewBool(false),
		},
		&ast.Ident{Name: "false"},
	),
}

func (check *Checker) defPrimitives() {
	for _, primitive := range primitives {
		defined := check.module.scope.Define(primitive)
		assert.Ok(defined == nil)
	}
}
