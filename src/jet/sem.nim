import
  std/strformat,

  jet/ast,
  jet/symbol,
  jet/module,

  lib/utils,
  lib/line_info

{.push, raises: [].}

type
  SemanticError* = object of CatchableError
    info* : LineInfo

template raiseSemanticError*(message: string; lineInfo: LineInfo) =
  raise (ref SemanticError)(msg: message, info: lineInfo)

func isSymDecl(tree: AstNode): bool =
  result =
    tree.kind == Branch and
    tree.branchKind in {ValDecl, VarDecl}

proc checkOperandTypes(module: ModuleRef; opNode: AstNode; left, right: TypeRef): TypeRef
  {.raises: [SemanticError].} =
  assert(opNode.kind == Operator)
  assert(left != nil and right != nil)

  let op = opNode.op

  result = case op:
    of OpAdd, OpSub, OpMul:
      if left.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator", opNode.info)

      if right.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator", opNode.info)

      if left.kind != right.kind:
        raiseSemanticError(
          "both expressions must be of the same type, got '" &
          $left & "' and '" & $right & "'",
          opNode.info)

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
  {.raises: [SemanticError, ModuleError].} =
  result = case expr.kind:
    of Empty:
      nil
    of Id:
      var sym = module.getSym(expr.id)

      if sym == nil:
        raiseSemanticError("unbound identifier: '" & expr.id & "'", expr.info)

      if sym.kind != skType and sym.`type` == nil:
        raiseSemanticError("expression '" & expr.id & "' has no type", expr.info)

      sym.`type`
    of Lit:
      case expr.lit.kind
      of lkInt:
        if expectedType != nil and
           expectedType.kind in {tyI8, tyI16, tyI32, tyI64}:
          # TODO: check int range
          return expectedType
        module.getBuiltinType(tyI32)
      of lkNil:
        # if expectedType != nil and
        #    expectedType.kind == tyRef:
        #     return expectedType
        module.getBuiltinType(tyNil)
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
          raiseSemanticError("expected typedesc, got expression", expr.info)
        typeDesc
      else:
        todo($expr.branchKind)

func genSym(module: ModuleRef; tree: AstNode): SymbolRef
  {.raises: [SemanticError, ValueError, ModuleError].} =
  result = nil

  if not tree.isSymDecl():
    return

  case tree.branchKind
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
      # check body type
      if not isCompatibleTypes(`type`, bodyType):
        raiseSemanticError(&"invalid type for '{id}'; expected {`type`}, got {bodyType}", idNode.info)

    if `type` == nil:
      raiseSemanticError(&"unable to infer type for '{id}'", idNode.info)

    let symKind = case tree.branchKind:
      of VarDecl: skVar
      of ValDecl: skVal
      else: unreachable()

    if `type`.kind == tyNil:
      raiseSemanticError("variable cannot be of type nil", idNode.info)

    result = SymbolRef(
      id: id,
      kind: symKind,
      `type`: `type`,
      scope: module.rootScope, # recursive
    )
  else:
    unimplemented()

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
