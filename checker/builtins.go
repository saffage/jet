package checker

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type BuiltInFunc func(node *ast.ParenList, args []*TypedValue) (*TypedValue, error)

type BuiltIn struct {
	name string
	f    BuiltInFunc
	t    *types.Func
}

func (b *BuiltIn) Owner() *Scope     { return nil }
func (b *BuiltIn) Type() types.Type  { return b.t }
func (b *BuiltIn) Name() string      { return b.name }
func (b *BuiltIn) Ident() *ast.Ident { return nil }
func (b *BuiltIn) Node() ast.Node    { return nil }

var builtIns = []*BuiltIn{
	{
		name: "magic",
		f:    builtInMagic,
		t: types.NewFunc(
			types.NewTuple(types.AnyTypeDesc),
			types.NewTuple(types.UntypedString),
			false,
		),
	},
	{
		name: "type_of",
		f:    builtInTypeOf,
		t: types.NewFunc(
			types.NewTuple(types.AnyTypeDesc),
			types.NewTuple(types.Any),
			false,
		),
	},
	{
		name: "print",
		f:    builtInPrint,
		t: types.NewFunc(
			types.Unit,
			types.NewTuple(types.Any),
			false,
		),
	},
	{
		name: "assert",
		f:    builtInAssert,
		t: types.NewFunc(
			types.Unit,
			types.NewTuple(types.Bool),
			false,
		),
	},
	{
		name: "asPtr",
		f:    builtInAsPtr,
		t: types.NewFunc(
			types.NewTuple(types.NewRef(types.U8)),
			types.NewTuple(types.String),
			false,
		),
	},
	{
		name: "as",
		f:    builtInAs,
		t: types.NewFunc(
			types.NewTuple(types.Any),
			types.NewTuple(types.AnyTypeDesc, types.Any),
			false,
		),
	},
	{
		name: "sizeOf",
		f:    builtInSizeOf,
		t: types.NewFunc(
			types.NewTuple(types.U64),
			types.NewTuple(types.AnyTypeDesc),
			false,
		),
	},
	{
		name: "emit",
		f:    builtInEmit,
		t: types.NewFunc(
			types.NewTuple(types.Unit),
			types.NewTuple(types.UntypedString),
			false,
		),
	},
}
