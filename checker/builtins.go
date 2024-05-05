package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type BuiltInFn func(args ast.Node, scope *Scope) (*Value, error)

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

func (b *BuiltIn) setType(t types.Type) { panic("can't change the type of the built-in function") }

var builtIns []*BuiltIn

func init() {
	builtIns = []*BuiltIn{
		{
			name: "magic",
			f:    builtInMagic,
			t: types.NewFunc(
				types.NewTuple(types.Primitives[types.AnyTypeDesc]),
				types.NewTuple(types.Primitives[types.UntypedString]),
			),
		},
		{
			name: "type_of",
			f:    builtInTypeOf,
			t: types.NewFunc(
				types.NewTuple(types.Primitives[types.AnyTypeDesc]),
				types.NewTuple(types.Primitives[types.Any]),
			),
		},
		{
			name: "print",
			f:    builtInPrint,
			t: types.NewFunc(
				types.Unit,
				types.NewTuple(types.Primitives[types.Any]),
			),
		},
	}
}
