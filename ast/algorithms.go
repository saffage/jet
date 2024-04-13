package ast

func UnwrapParen(node Node) Node {
	n, ok := node.(*ParenExpr)

	for ok {
		node = n.X
		n, ok = node.(*ParenExpr)
	}

	return node
}
