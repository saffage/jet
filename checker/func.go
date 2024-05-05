package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

type Func struct {
	owner *Scope
	scope *Scope
	t     *types.Func
	node  *ast.FuncDecl
}

func NewFunc(owner *Scope, t *types.Func, node *ast.FuncDecl) *Func {
	return &Func{owner, NewScope(owner), t, node}
}

func (sym *Func) Owner() *Scope     { return sym.owner }
func (sym *Func) Type() types.Type  { return sym.t }
func (sym *Func) Name() string      { return sym.node.Name.Name }
func (sym *Func) Ident() *ast.Ident { return sym.node.Name }
func (sym *Func) Node() ast.Node    { return sym.node }

func (sym *Func) setType(t types.Type) {
	if t, _ := t.(*types.Func); t != nil {
		sym.t = t
		return
	}

	panic(fmt.Sprintf("type '%s' is not a function", t))
}

// func (sym *Func) Check(args []types.Type) error {
// 	if len(args) > len(t.Params) {
// 		return NewErrorf(node, "too many arguments (expected %d)", len(b.params))
// 	}

// 	if len(args) < len(t.Params) {
// 		return NewErrorf(node, "not enough arguments (expected %d)", len(b.params))
// 	}

// 	for i := 0; i < maxlen; i++ {
// 		if i < len(args.Nodes) {
// 			if type_, err := scope.TypeOf(args.Nodes[i]); err == nil {
// 				actual = type_
// 				node = args.Nodes[i]
// 			} else {
// 				return err
// 			}
// 		}

// 		if i < len(b.params) {
// 			expected = b.params[i]
// 		}

// 		if expected == nil {
// 		}

// 		if actual == nil {
// 		}

// 		if !expected.Equals(actual) {
// 			return NewErrorf(node, "expected '%s' for %d argument but got '%s'", expected, i+1, actual)
// 		}
// 	}

// 	return nil
// }
