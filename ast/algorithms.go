package ast

func UnwrapParen(node Node) Node {
	n, ok := node.(*ParenList)

	for ok && len(n.Exprs) == 1 {
		node = n.Exprs[0]
		n, ok = node.(*ParenList)
	}

	return node
}
