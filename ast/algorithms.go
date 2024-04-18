package ast

func UnwrapParenExpr(node Node) Node {
	n, ok := node.(*ParenList)

	for ok && len(n.Nodes) == 1 {
		node = n.Nodes[0]
		n, ok = node.(*ParenList)
	}

	return node
}
