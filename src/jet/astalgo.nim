import
  jet/ast

func printTreeAux(tree: AstNode; buffer: var string; indent: string; last: bool) =
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

  buffer &= indent & (if last: lastLeaf else: leaf)
  buffer &= $tree & "\n"

  if not tree.isLeaf():
    let indent = indent & (if last: space else: connector)

    for i, node in tree.children:
      printTreeAux(node, buffer, indent, i == tree.children.high)

proc printTree*(tree: AstNode) =
  var buffer = newStringOfCap(512) # random number
  printTreeAux(tree, buffer, "", true)
  echo(buffer)

func treeRepr*(tree: AstNode): string =
  result = newStringOfCap(512) # random number
  printTreeAux(tree, result, "", true)
