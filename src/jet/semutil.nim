# NOTE: temporary file

import
  jet/ast

func isAnnotation*(tree: AstNode): bool =
  result =
    tree.kind == Branch and
    tree.branchKind == ExprRound and
    tree.children[0].kind == Branch and
    tree.children[0].branchKind == Prefix and
    tree.children[0].children[0].kind == Operator and
    tree.children[0].children[0].op == OpAnnot

func getAnnotationName*(tree: AstNode): string =
  assert(tree.isAnnotation())
  result = tree.children[0].children[1].id

func getAnnotationArgs*(tree: AstNode): seq[AstNode] =
  assert(tree.isAnnotation())
  result = tree.children[1].children
