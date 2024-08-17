package ast

import "fmt"

// Returns a data, stored in the node.
//
// Type of the node must be one [Name], [Type] or [Underscore].
func Data(node Node) string {
	switch node := node.(type) {
	case *Name:
		return node.Data

	case *Type:
		return node.Data

	case *Underscore:
		return node.Data

	default:
		panic(fmt.Sprintf("ast.Data: cannot get data from the node of type '%T'", node))
	}
}
