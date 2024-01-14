import
  std/strformat,

  jet/literal

type
  AstNodeKind* = enum
    Empty
    Id
    Lit
    Branch

  AstNodeBranchKind* = enum
    Func
    VarDecl
    Tuple   ## (a, b)
    Block   ## (a; b)
    Infix   ## a ~ b

  AstNodeRef* = ref AstNode
  AstNode* {.byref.} = object
    case kind* : AstNodeKind
    of Empty:
      nil
    of Id:
      id* : string
    of Lit:
      lit* : Literal
    of Branch:
      branchKind* : AstNodeBranchKind
      children*   : seq[AstNode]

func isLeaf*(tree: AstNode): bool =
  tree.kind != Branch

func len*(tree: AstNode): int =
  if tree.isLeaf(): 0
  else: tree.children.len()

func `$`*(tree: AstNode): string =
  result = $tree.kind

  case tree.kind
  of Id:
    result &= &" \"{tree.id}\""
  of Lit:
    result &= $tree.lit
  else:
    discard
