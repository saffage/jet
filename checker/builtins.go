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
		name: "builtin",
		f:    builtInBuiltin,
		t: types.NewFunc(
			types.NewTuple(types.UntypedString),
			types.NewTuple(types.AnyTypeDesc),
			nil,
		),
	},
	{
		name: "type_of",
		f:    builtInTypeOf,
		t: types.NewFunc(
			types.NewTuple(types.Any),
			types.NewTuple(types.AnyTypeDesc),
			nil,
		),
	},
	{
		name: "print",
		f:    builtInPrint,
		t: types.NewFunc(
			types.NewTuple(types.Any),
			types.Unit,
			nil,
		),
	},
	{
		name: "println",
		f:    builtInPrint, // why not
		t: types.NewFunc(
			types.NewTuple(types.Any),
			types.Unit,
			nil,
		),
	},
	{
		name: "assert",
		f:    builtInAssert,
		t: types.NewFunc(
			types.NewTuple(types.Bool),
			types.Unit,
			nil,
		),
	},
	{
		name: "as_ptr",
		f:    builtInAsPtr,
		t: types.NewFunc(
			types.NewTuple(types.String),
			types.NewTuple(types.NewRef(types.U8)),
			nil,
		),
	},
	{
		name: "cast",
		f:    builtInCast,
		t: types.NewFunc(
			types.NewTuple(types.AnyTypeDesc, types.Any),
			types.NewTuple(types.Any),
			nil,
		),
	},
	{
		name: "size_of",
		f:    builtInSizeOf,
		t: types.NewFunc(
			types.NewTuple(types.AnyTypeDesc),
			types.NewTuple(types.U64),
			nil,
		),
	},
	{
		name: "emit",
		f:    builtInEmit,
		t: types.NewFunc(
			types.NewTuple(types.UntypedString),
			types.NewTuple(types.Unit),
			nil,
		),
	},
}
