import
  jet/ast,
  jet/literal,

  lib/utils

proc ast2c*(buf: var string; tree: AstNode; isStmt = false; indentation = "") =
  buf &= indentation

  case tree.kind
  of Id:
    buf &= tree.id
  of Lit:
    buf &= tree.lit.pretty()
  of Branch:
    case tree.branchKind
    of ExprRound:
      buf.ast2c(tree.children[0])
      buf &= '('
      for i, node in tree.children[1].children:
        if i > 0: buf &= ", "
        buf.ast2c(node)
      buf &= ')'
    of VarDecl, ValDecl:
      if tree.branchKind == ValDecl:
        buf &= "const "
      buf &= "jet_"
      buf.ast2c(tree.children[1])
      buf &= " "
      buf.ast2c(tree.children[0])
      buf &= " = "
      buf.ast2c(tree.children[2])
    else:
      unimplemented("Branch " & $tree.branchKind)
  else:
    unimplemented($tree.kind)

  if isStmt:
    buf &= ";\n"

proc ast2c*(tree: AstNode): string =
  assert(tree.kind == Branch)
  assert(tree.branchKind == Block)

  result = """
#include "lib/core/cgen/jet.h"

#include <stdlib.h>
#include <stdio.h>

void JetMain(void) {
"""
  for node in tree.children:
    result.ast2c(node, true, "  ")
  result &= """
}

int main(int argc, char **argv) {
  JetMain();

  return 0;
}
"""
