import
  std/options,

  jet/literal,

  lib/lineinfo

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
    RecFunc
    Func
    If
    IfBranch
    ElseBranch
    While
    Return
    ValDecl
    VarDecl
    List        ## (a, b)
    Block       ## (a; b)
    Infix       ## a ~ b
    Prefix      ## ~a
    Postfix     ## a~
    ExprCurly   ## a{...}
    ExprRound   ## a(...)

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
    rng* : FileRange

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
    OpDivInt = "" # TODO: integer division operator
    OpMod    = "%"
    OpShl    = "<<"
    OpShr    = ">>"
    OpRef    = "&"
    OpRefVar = "&var"
    OpDollar = "$"
    OpDot    = "."
    OpAnnot  = "@"

  OperatorNotation* = enum
    Infix
    Prefix
    Postfix

func toOperatorKind*(value: string): Option[OperatorKind] =
  result = none(OperatorKind)
  
  for kind in OperatorKind:
    if $kind == value:
      result = some(kind)
      break

func notation*(kind: OperatorKind): set[OperatorNotation] =
  result = case kind:
    of OpNot: {Prefix}
    of OpAnd: {Infix}
    of OpOr: {Infix}
    of OpEq: {Infix}
    of OpNe: {Infix}
    of OpLt: {Infix}
    of OpLe: {Infix}
    of OpGt: {Infix}
    of OpGe: {Infix}
    of OpAdd: {Infix}
    of OpSub: {Infix}
    of OpMul: {Infix}
    of OpDiv: {Infix}
    of OpDivInt: {Infix}
    of OpMod: {Infix}
    of OpShl: {Infix}
    of OpShr: {Infix}
    of OpRef: {Prefix}
    of OpRefVar: {Prefix}
    of OpDollar: {Prefix}
    of OpDot: {Infix}
    of OpAnnot: {Prefix}

func isLeaf*(tree: AstNode): bool =
  result = tree.kind != Branch

func len*(tree: AstNode): int =
  result =
    if tree.isLeaf():
      0
    else:
      tree.children.len()

func `$`*(tree: AstNode): string =
  result =
    if tree.kind != Branch:
      $tree.kind
    else:
      $tree.branchKind

  if tree.rng != emptyFileRange:
    result &= "[" & $tree.rng & "]"

  case tree.kind
  of Id:
    result &= " = `" & tree.id & "`"
  of Lit:
    result &= " = " & tree.lit.pretty()
  of Operator:
    result &= " = " & $tree.op
  else:
    discard

func initAstNode*(kind: AstNodeKind; rng = emptyFileRange): AstNode =
  result = AstNode(kind: kind, rng: rng)

func initAstNodeEmpty*(rng = emptyFileRange): AstNode =
  result = AstNode(kind: Empty, rng: rng)

func initAstNodeId*(id: sink string; rng = emptyFileRange): AstNode =
  result = AstNode(kind: Id, id: id, rng: rng)

func initAstNodeLit*(lit: Literal; rng = emptyFileRange): AstNode =
  result = AstNode(kind: Lit, lit: lit, rng: rng)

func initAstNodeOperator*(op: OperatorKind; rng = emptyFileRange): AstNode =
  result = AstNode(kind: Operator, op: op, rng: rng)

func initAstNodeBranch*(branchKind: AstNodeBranchKind; children = newSeq[AstNode](); rng = emptyFileRange): AstNode =
  result = AstNode(kind: Branch, branchKind: branchKind, children: children, rng: rng)
