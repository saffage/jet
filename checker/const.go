package checker

// type Const struct {
// 	owner *Scope
// 	node  *ast.GenericDecl
// 	name  *ast.Ident
// 	t     types.Type
// 	value constant.Value
// }

// func NewConst(owner *Scope, node *ast.GenericDecl, name *ast.Ident) (*Const, error) {
// 	if node.Kind != ast.ConstDecl {
// 		return nil, NewError(node, "expected constant declaration")
// 	}

// 	if node.Field.Value == nil {
// 		return nil, NewError(name, "value is required for constant")
// 	}

// 	value := constant.FromNode(node.Field.Value)
// 	sym := &Const{
// 		owner: owner,
// 		// type_: types.FromConstant(value.Kind()),
// 		value: value,
// 		node:  node,
// 		name:  name,
// 	}

// 	return sym, nil
// }

// func (v *Const) Owner() *Scope     { return v.owner }
// func (v *Const) Type() types.Type  { return v.t }
// func (v *Const) Name() string      { return v.name.Name }
// func (v *Const) Ident() *ast.Ident { return v.name }
// func (v *Const) Node() ast.Node    { return v.node }

// func (v *Const) setType(t types.Type) { v.t = t }
