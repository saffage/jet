import
  std/strformat,
  std/options,

  jet/literal

type
  AstNodeKind* = enum
    Empty
    Id
    Lit
    Operator
    Branch

  AstNodeBranchKind* = enum
    Func
    If
    IfBranch
    ElseBranch
    While
    Return
    ValDecl
    VarDecl
    Tuple   ## (a, b)
    Block   ## (a; b)
    Infix   ## a ~ b
    Prefix  ## ~a
    Postfix ## a~

  AstNodeRef* = ref AstNode
  AstNode* {.byref.} = object
    case kind* : AstNodeKind
    of Empty:
      nil
    of Id:
      id* : string
    of Lit:
      lit* : Literal
    of Operator:
      op* : OperatorKind
    of Branch:
      branchKind* : AstNodeBranchKind
      children*   : seq[AstNode]

  OperatorKind* = enum
    OpNot    = "not"
    OpAnd    = "and"
    OpOr     = "or"
    OpEq     = "=="
    OpNe     = "!="
    OpLt     = "<"
    OpLe     = "<="
    OpGt     = ">"
    OpGe     = ">="
    OpAdd    = "+"
    OpSub    = "-"
    OpMul    = "*"
    OpDiv    = "/"
    OpDivInt = ""
    OpMod    = "%"
    OpShl    = "<<"
    OpShr    = ">>"

const
  emptyNode* = AstNode(kind: Empty)

func toOperatorKind*(value: string): Option[OperatorKind] =
  result = none(OperatorKind)
  for kind in OperatorKind:
    if $kind == value:
      result = some(kind)
      break

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
    result &= &" {tree.lit.pretty()}"
  of Operator:
    result &= &" {tree.op}"
  else:
    discard
