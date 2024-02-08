import
  jet/ast

{.push, raises: [].}

func printTreeAux(tree: AstNode; buffer: var string; indent: string; isLast: bool) =
  when defined(jetAstAsciiRepr):
    const connector = "|  "
    const leaf      = "|- "
    const lastLeaf  = "'- "
    const space     = "   "
  else:
    const connector = "│  "
    const leaf      = "├─╴"
    const lastLeaf  = "└─╴"
    const space     = "   "

  buffer &= indent & (if isLast: lastLeaf else: leaf)
  buffer &= $tree & "\n"

  if not tree.isLeaf():
    let indent = indent & (if isLast: space else: connector)

    for i, node in tree.children:
      printTreeAux(node, buffer, indent, i == tree.children.high)

func treeRepr*(tree: AstNode): string =
  result = newStringOfCap(512) # random number
  printTreeAux(tree, result, "", true)

proc printTree*(tree: AstNode) =
  echo(tree.treeRepr())

{.pop.} # raises: []
