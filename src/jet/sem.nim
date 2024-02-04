import
  std/strformat,
  std/strutils,
  std/sequtils,
  std/options,

  jet/ast,
  jet/symbol,
  jet/module,
  jet/magics,
  jet/semutil,

  lib/utils,
  lib/lineinfo

{.push, raises: [].}

type
  SemanticError* = object of CatchableError
    rng* : FileRange

template raiseSemanticError*(message: string; node: AstNode) =
  raise (ref SemanticError)(msg: message, rng: node.rng)

template raiseSemanticError*(message: string; fileRange: FileRange) =
  raise (ref SemanticError)(msg: message, rng: fileRange)

template raiseSemanticError*(message: string; filePos: FilePosition) =
  raise (ref SemanticError)(msg: message, rng: filePos.withLength(0))

func isSymDecl(tree: AstNode): bool =
  result =
    tree.kind == Branch and
    tree.branchKind in {ValDecl, VarDecl, Type}

proc checkOperandTypes(module: ModuleRef; opNode: AstNode; left, right: TypeRef): TypeRef
  {.raises: [SemanticError].} =
  assert(opNode.kind == Operator)
  assert(left != nil and right != nil)

  let op = opNode.op

  result = case op:
    of OpAdd, OpSub, OpMul:
      if left.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator", opNode.rng)

      if right.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator", opNode.rng)

      if left.kind != right.kind:
        raiseSemanticError(
          "both expressions must be of the same type, got '" &
          $left & "' and '" & $right & "'",
          opNode.rng)

      left
    else:
      todo()

proc getTypeDesc(module: ModuleRef; expr: AstNode): TypeRef =
  result = nil

  case expr.kind
  of Id:
    let sym = module.getSym(expr.id)
    if sym.kind == skType:
      return sym.`type`
  of Branch:
    case expr.branchKind:
    of Prefix:
      if expr.children[0].kind == Operator and
          expr.children[0].op == OpRef:
        return TypeRef(kind: tyRef, parent: module.getTypeDesc(expr.children[1]))
    else:
      return
  else:
    return

proc typeOfExpr(module: ModuleRef; expr: AstNode; expectedType = nil.TypeRef): TypeRef
  {.raises: [SemanticError].} =
  result = case expr.kind:
    of Empty:
      nil
    of Id:
      var sym = module.getSym(expr.id)

      if sym == nil:
        raiseSemanticError("unbound identifier: '" & expr.id & "'", expr)

      if sym.kind != skType and sym.`type` == nil:
        raiseSemanticError("expression '" & expr.id & "' has no type", expr)

      sym.`type`
    of Lit:
      case expr.lit.kind
      of lkInt:
        if expectedType != nil and
           expectedType.kind in {tyI8, tyI16, tyI32, tyI64}:
          # TODO: check int range
          return expectedType
        module.getMagicSym(mTypeI32).`type`
      of lkNil:
        unimplemented("nil")
      else:
        todo($expr.lit.kind)
    of Operator:
      todo()
    of Branch:
      case expr.branchKind
      of Infix:
        let opNode = expr.children[0]
        let left = module.typeOfExpr(expr.children[1])
        let right = module.typeOfExpr(expr.children[2])
        module.checkOperandTypes(opNode, left, right)
      of Prefix:
        let opKind = expr.children[0].op
        let operand = module.typeOfExpr(expr.children[1])

        if opKind == OpRef:
          TypeRef(kind: tyRef, parent: operand)
        else:
          todo()
      of ExprCurly:
        let typeDesc = module.getTypeDesc(expr.children[0])
        if typeDesc == nil:
          raiseSemanticError("expected typedesc, got expression", expr)
        typeDesc
      else:
        todo($expr.branchKind)

proc genSym(module: ModuleRef; tree: AstNode): SymbolRef
  {.raises: [SemanticError, ValueError].} =
  result = nil

  if not tree.isSymDecl():
    return

  case tree.branchKind
  of Type:
    let name = tree.children[0]
    if name.kind != Id:
      unimplemented("name is not Id")

    let body = tree.children[1]

    case body.kind:
    of Branch:
      case body.branchKind:
      of ExprRound:
        if not body.isAnnotation(): unimplemented()
        if body.getAnnotationName() != "Magic": todo()

        let args = body.getAnnotationArgs()
        if args.len() != 1: todo()

        let arg = args[0].lit.stringVal
        let magic = try:
          parseEnum[MagicKind]('m' & arg)
        except ValueError:
          raiseSemanticError("unknown magic: '" & arg & "'", args[0].rng)
        let magicSym = module.getMagicSym(magic)

        magic.markAsResolved()

        result = SymbolRef(
          id: name.id,
          kind: skType,
          `type`: magicSym.`type`,
          scope: nil, # idk
          magic: some(magic),
        )
      else:
        unimplemented()
    else:
      unimplemented()

  of VarDecl, ValDecl:
    let idNode = tree.children[0]
    let id = idNode.id
    let typeExpr = tree.children[1]
    let body = tree.children[2]

    var `type` = module.typeOfExpr(typeExpr)
    let bodyType = module.typeOfExpr(body, `type`)

    if `type` == nil:
      `type` = module.typeOfExpr(body)
    else:
      if bodyType != nil and not isCompatibleTypes(`type`, bodyType):
        raiseSemanticError(&"invalid type for '{id}'; expected {`type`}, got {bodyType}", idNode)

    if `type` == nil:
      raiseSemanticError(&"unable to infer type for '{id}'", idNode)

    let symKind = case tree.branchKind:
      of VarDecl: skVar
      of ValDecl: skVal
      else: unreachable()

    if `type`.kind == tyNil:
      raiseSemanticError("variable cannot be of type nil", idNode)

    result = SymbolRef(
      id: id,
      kind: symKind,
      `type`: `type`,
      scope: module.rootScope, # recursive
    )
  else:
    unimplemented()

proc assertMagicsResolved(module: ModuleRef)
  {.raises: [SemanticError, ValueError].} =
  let unresolvedMagics = getUnresolvedMagics()

  if unresolvedMagics != {}:
    let magicsAsStr = unresolvedMagics.toSeq().join(", ")
    raiseSemanticError(
      &"the following magics was not resolved: {magicsAsStr}",
      emptyFilePos)

proc traverseSymbols*(module: ModuleRef; rootTree: AstNode)
  {.raises: [ModuleError, SemanticError, ValueError].} =
  for tree in rootTree.children:
    if tree.kind == Branch and tree.branchKind == Block:
      module.traverseSymbols(tree)
    else:
      let sym = module.genSym(tree)

      if sym != nil:
        module.registerSymbol(sym)

proc traverseSymbols*(module: ModuleRef)
  {.raises: [ModuleError, SemanticError, ValueError].} =
  module.traverseSymbols(module.rootTree)
  module.assertMagicsResolved() # TODO: it's not supposed to be here

{.pop.} # raises: []
