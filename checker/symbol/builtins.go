package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
)

// TODO replace with generic binding.
type BuiltinParam struct {
	name  *ast.Ident
	type_ types.Type
}

// Returns:
//   - constant.Value
//   - ast.Node
//   - symbol.Symbol
//   - types.Type
type BuiltinCallFn func(b *Builtin, scope Scope, call *ast.BuiltInCall) any

type Builtin struct {
	base
	params []BuiltinParam
	fn     BuiltinCallFn
}

func NewBuiltin(id ID, node *ast.BuiltInCall) *Builtin {
	return &Builtin{
		base: base{
			id:   id,
			name: node.Name,
			node: node,
		},
	}
}

func builtinMagic(b *Builtin, scope Scope, call *ast.BuiltInCall) any {
	args, ok := call.X.(*ast.ParenList)
	if !ok {
		panic(NewError(call.X, "expected argument list"))
	}

	err := b.checkArgTypes(scope, args)
	if err != nil {
		return err
	}

	arg1, ok := args.Nodes[0].(*ast.Literal)
	if !ok {
		panic(NewError(args.Nodes[0], "expected literal"))
	}

	magicName := arg1.Value

	switch magicName {
	case "Bool":
		return types.TypeDesc{Type: types.Bool{}}

	case "I32":
		return types.TypeDesc{Type: types.I32{}}

	default:
		panic(NewErrorf(call, "unknown magic '%s'", magicName))
	}
}

func builtinTypeOf(b *Builtin, scope Scope, call *ast.BuiltInCall) any {
	args, ok := call.X.(*ast.ParenList)
	if !ok {
		panic(NewError(call, "expected argument list"))
	}

	err := b.checkArgTypes(scope, args)
	if err != nil {
		return err
	}

	arg1 := args.Nodes[0]

	type_, err := TypeOf(scope, arg1)
	if err != nil {
		return err
	}

	return types.Type(types.TypeDesc{types.TypedFromUntyped(type_)})
}

func (b *Builtin) checkArgTypes(scope Scope, args *ast.ParenList) error {
	maxlen := max(len(args.Nodes), len(b.params))

	for i := 0; i < maxlen; i++ {
		var expected, actual types.Type
		var node ast.Node = args

		if i < len(args.Nodes) {
			if type_, err := TypeOf(scope, args.Nodes[i]); err == nil {
				actual = type_
				node = args.Nodes[i]
			} else {
				return err
			}
		}

		if i < len(b.params) {
			expected = b.params[i].type_
		}

		assert.Ok(expected != nil || actual != nil)

		if expected == nil {
			return NewErrorf(node, "too many arguments (expected %d)", len(b.params))
		}

		if actual == nil {
			return NewErrorf(node, "not enough arguments (expected %d)", len(b.params))
		}

		if !expected.Equals(actual) {
			return NewErrorf(node, "expected '%s' for %d argument but got '%s'", expected, i+1, actual)
		}
	}

	return nil
}
