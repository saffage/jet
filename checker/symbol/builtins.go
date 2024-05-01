package symbol

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker/types"
	"github.com/saffage/jet/internal/assert"
)

// TODO replace with generic binding.
type BuiltInParam struct {
	name  *ast.Ident
	type_ types.Type
}

// Returns:
//   - constant.Value
//   - ast.Node
//   - symbol.Symbol
//   - types.Type
type BuiltInCallFn func(b *BuiltIn, scope Scope, call *ast.BuiltInCall) any

type BuiltIn struct {
	id    ID
	owner Scope

	name   string
	fn     BuiltInCallFn
	params []BuiltInParam
}

func (v *BuiltIn) ID() ID            { return v.id }
func (v *BuiltIn) Owner() Scope      { return v.owner }
func (v *BuiltIn) Type() types.Type  { return nil }
func (v *BuiltIn) Name() string      { return v.name }
func (v *BuiltIn) Ident() *ast.Ident { return nil }
func (v *BuiltIn) Node() ast.Node    { return nil }

func (v *BuiltIn) setType(t types.Type) { panic("built-in functions have no type") }

func builtInMagic(b *BuiltIn, scope Scope, call *ast.BuiltInCall) any {
	args, ok := call.X.(*ast.ParenList)
	if !ok {
		return NewError(call.X, "expected argument list")
	}

	err := b.checkArgTypes(scope, args)
	if err != nil {
		return err
	}

	arg1, ok := args.Nodes[0].(*ast.Literal)
	if !ok {
		return NewError(args.Nodes[0], "expected literal")
	}

	magicName := arg1.Value

	switch magicName {
	case "Bool":
		return types.TypeDesc{Type: types.Bool{}}

	case "I32":
		return types.TypeDesc{Type: types.I32{}}

	default:
		return NewErrorf(call, "unknown magic '%s'", magicName)
	}
}

func builtInTypeOf(b *BuiltIn, scope Scope, call *ast.BuiltInCall) any {
	args, ok := call.X.(*ast.ParenList)
	if !ok {
		return NewError(call, "expected argument list")
	}

	err := b.checkArgTypes(scope, args)
	if err != nil {
		return err
	}

	arg1 := args.Nodes[0]

	t, err := TypeOf(scope, arg1)
	if err != nil {
		return err
	}

	typedesc := types.TypeDesc{
		Type: types.TypedFromUntyped(t),
	}

	return types.Type(typedesc)
}

func (b *BuiltIn) checkArgTypes(scope Scope, args *ast.ParenList) error {
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

var builtIns []*BuiltIn

func init() {
	builtIns = []*BuiltIn{
		{
			id:    nextID(),
			owner: nil,
			name:  "magic",
			fn:    builtInMagic,
			params: []BuiltInParam{
				{
					name:  &ast.Ident{Name: "name"},
					type_: types.UntypedString{},
				},
			},
		},
		{
			id:    nextID(),
			owner: nil,
			name:  "type_of",
			fn:    builtInTypeOf,
			params: []BuiltInParam{
				{
					name:  &ast.Ident{Name: "expr"},
					type_: types.Any{},
				},
			},
		},
	}
}
