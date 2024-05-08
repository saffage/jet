package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type BuiltInFn func(node *ast.ParenList, args []*TypedValue) *TypedValue

type BuiltIn struct {
	name string
	f    BuiltInFn
	t    *types.Func
}

func (b *BuiltIn) Owner() *Scope     { return nil }
func (b *BuiltIn) Type() types.Type  { return b.t }
func (b *BuiltIn) Name() string      { return b.name }
func (b *BuiltIn) Ident() *ast.Ident { return nil }
func (b *BuiltIn) Node() ast.Node    { return nil }

func (check *Checker) defBuiltIns() {
	check.builtIns = []*BuiltIn{
		{
			name: "magic",
			f:    check.builtInMagic,
			t: types.NewFunc(
				types.NewTuple(types.AnyTypeDesc),
				types.NewTuple(types.UntypedString),
			),
		},
		{
			name: "type_of",
			f:    check.builtInTypeOf,
			t: types.NewFunc(
				types.NewTuple(types.AnyTypeDesc),
				types.NewTuple(types.Any),
			),
		},
		{
			name: "print",
			f:    check.builtInPrint,
			t: types.NewFunc(
				types.Unit,
				types.NewTuple(types.Any),
			),
		},
		{
			name: "assert",
			f:    check.builtInAssert,
			t: types.NewFunc(
				types.Unit,
				types.NewTuple(types.Bool),
			),
		},
	}
}
