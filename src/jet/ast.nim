import
  std/strformat,
  std/options,

  jet/literal,

  lib/line_info

type
  AstNodeKind* = enum
    Empty
    Id
    Lit
    Operator
    Branch

  AstNodeBranchKind* = enum
    Struct
    Enum
    Type
    Func
    If
    IfBranch
    ElseBranch
    While
    Return
    ValDecl
    VarDecl
    List   ## (a, b)
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
    info* : LineInfo

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

  if tree.info != LineInfo():
    result &= &"[{tree.info}]"

  case tree.kind
  of Id:
    result &= &" = \"{tree.id}\""
  of Lit:
    result &= &" = {tree.lit.pretty()}"
  of Operator:
    result &= &" = {tree.op}"
  else:
    discard

func initAstNode*(kind: AstNodeKind; info = LineInfo()): AstNode =
  result = AstNode(kind: kind, info: info)

func initAstNodeEmpty*(info = LineInfo()): AstNode =
  result = AstNode(kind: Empty, info: info)

func initAstNodeId*(id: sink string; info = LineInfo()): AstNode =
  result = AstNode(kind: Id, id: id, info: info)

func initAstNodeLit*(lit: Literal; info = LineInfo()): AstNode =
  result = AstNode(kind: Lit, lit: lit, info: info)

func initAstNodeOperator*(op: OperatorKind; info = LineInfo()): AstNode =
  result = AstNode(kind: Operator, op: op, info: info)

func initAstNodeBranch*(branchKind: AstNodeBranchKind; children = newSeq[AstNode](); info = LineInfo()): AstNode =
  result = AstNode(kind: Branch, branchKind: branchKind, children: children, info: info)
