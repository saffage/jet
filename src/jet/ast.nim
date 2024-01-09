import std/sequtils

import ast/nodes

export nodes


# ----- Leaf ------ #
proc newProgram*(): Node =
    result = newNode(nkProgram)


# ----- Stmt ------ #
proc newFunc*(name, params, returnType, body: Node): Node =
    assert(name != nil and name.kind == nkId)
    assert(params != nil and params.kind == nkParen)
    assert(returnType != nil)
    assert(body != nil and body.kind in {nkEmpty, nkExprList})

    result  = newNode(nkFunc)
    result &= name
    result &= params
    result &= returnType
    result &= body
    result &= newEmptyNode()

proc newType*(name, body: Node): Node =
    assert(name != nil)
    assert(body != nil and body.kind == nkExprList)

    result  = newNode(nkType)
    result &= name
    result &= body
    result &= newEmptyNode()

proc newReturnStmt*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkReturn)
    result &= expr


# ----- Expr ------ #
proc newEmptyIfExpr*(): Node =
    result = newNode(nkIf)

proc newIfExpr*(branches: openArray[Node]; elseBranch: Node = nil): Node =
    assert(branches.len() > 0)
    assert(branches.allIt(it != nil and it.kind == nkIfBranch), $branches.mapIt($it.kind))

    result  = newNode(nkIf)
    result &= branches

    if elseBranch != nil:
        assert(elseBranch.kind == nkElseBranch, $elseBranch.kind)
        result &= elseBranch

proc newIfExpr*(branch: Node; elseBranch: Node = nil): Node =
    result = newIfExpr([branch], elseBranch)

proc newIfBranch*(expr: Node; body: Node): Node =
    assert(expr != nil)
    assert(body != nil and body.kind == nkExprList, if body == nil: "nil" else: $body.kind)

    result  = newNode(nkIfBranch)
    result &= expr
    result &= body

proc newElseBranch*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkElseBranch)
    result &= expr

proc newEmptyMatchExpr*(): Node =
    result = newNode(nkMatch)

proc newMatchExpr*(expr: Node; cases: openArray[Node]; elseBranch: Node = nil): Node =
    assert(expr != nil)
    assert(cases.allIt(it != nil))

    result  = newNode(nkMatch)
    result &= expr
    result &= cases

    if elseBranch != nil:
        assert(elseBranch.kind == nkElseBranch, $elseBranch.kind)
        result &= elseBranch

proc newMatchExpr*(expr: Node; cases: Node; elseBranch: Node = nil): Node =
    result = newMatchExpr(expr, cases, elseBranch)

proc newMatchCase*(expr, bodyOrGuard: Node): Node =
    assert(expr != nil)
    assert(bodyOrGuard != nil and bodyOrGuard.kind in {nkIfBranch},
        if bodyOrGuard != nil: $bodyOrGuard.kind else: "nil")

    result  = newNode(nkCase)
    result &= expr
    result &= bodyOrGuard

proc newEmptyElseBranch*(): Node =
    result = newNode(nkElseBranch)

proc newExprEqExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprEqExpr)
    result &= left
    result &= right

proc newExprColonExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprColonExpr)
    result &= left
    result &= right

proc newExprDotExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprDotExpr)
    result &= left
    result &= right


# ----- Other ------ #
proc newEmptyBrace*(): Node =
    result = newNode(nkBrace)

proc newEmptyExprBrace*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprBrace)
    result &= expr
    result &= newEmptyBrace()

proc newExprBrace*(expr: Node; elems: openArray[Node]): Node =
    assert(elems.allIt(it != nil))

    result = newEmptyExprBrace(expr)
    result[1].children = @elems

proc newExprBrace*(expr: Node; elem: Node): Node =
    result = newExprBrace(expr, [elem])

proc newExprBracketFromBrace*(expr: Node; brace: Node): Node =
    assert(expr != nil)
    assert(brace != nil and brace.kind == nkBrace)

    result  = newNode(nkExprBrace)
    result &= expr
    result &= brace

proc newEmptyParen*(): Node =
    result = newNode(nkParen)

proc newEmptyExprParen*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprParen)
    result &= expr
    result &= newEmptyParen()

proc newExprParen*(expr: Node; elems: openArray[Node]): Node =
    assert(elems.allIt(it != nil))

    result = newEmptyExprParen(expr)
    result[1].children = @elems

proc newExprParen*(expr: Node; elem: Node): Node =
    result = newExprParen(expr, [elem])

proc newExprParenFromParen*(expr: Node; paren: Node): Node =
    assert(expr != nil)
    assert(paren != nil and paren.kind == nkParen)

    result  = newNode(nkExprParen)
    result &= expr
    result &= paren

proc newEmptyBracket*(): Node =
    result = newNode(nkBracket)

proc newEmptyExprBracket*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprBracket)
    result &= expr
    result &= newEmptyBracket()

proc newExprBracket*(expr: Node; elems: openArray[Node]): Node =
    assert(elems.allIt(it != nil))

    result = newEmptyExprBracket(expr)
    result[1].children = @elems

proc newExprBracket*(expr: Node; elem: Node): Node =
    result = newExprBracket(expr, [elem])

proc newExprBracketFromBracket*(expr: Node; bracket: Node): Node =
    assert(expr != nil)
    assert(bracket != nil and bracket.kind == nkBracket)

    result  = newNode(nkExprBracket)
    result &= expr
    result &= bracket


proc newPrefix*(op: Node; operand: Node): Node =
    assert(op != nil)
    assert(operand != nil)

    result  = Node(kind: nkPrefix)
    result &= op
    result &= operand

proc newPostfix*(op: Node; operand: Node): Node =
    assert(op != nil)
    assert(operand != nil)

    result  = Node(kind: nkPostfix)
    result &= op
    result &= operand

proc newInfix*(op: Node; leftOperand, rightOperand: Node): Node =
    assert(op != nil)
    assert(leftOperand != nil)
    assert(rightOperand != nil)

    result  = Node(kind: nkInfix)
    result &= op
    result &= leftOperand
    result &= rightOperand

proc newAnnotationList*(annotations: openArray[Node]): Node =
    result = newNode(nkAnnotationList)

    for annotation in annotations:
        assert(annotation != nil and annotation.kind in {nkId, nkExprParen, nkAnnotationList},
            if annotation == nil: "nil" else: $annotation.kind)

        if annotation.kind == nkAnnotationList:
            result &= annotation.children
        else:
            result &= annotation

proc newAnnotationList*(annotation: Node): Node =
    result = newAnnotationList([annotation])

proc newEmptyAnnotationList*(): Node =
    result = newAnnotationList([])

proc newVarDecl*(names: openArray[Node]; typeExpr: Node; expr: Node = nil): Node =
    assert(names.len() > 0)
    assert(names.allIt(it != nil and it.kind == nkId))

    result  = newNode(nkVarDecl)
    result &= names
    result &= (if typeExpr != nil: typeExpr else: newEmptyNode())
    result &= (if expr != nil: expr else: newEmptyNode())
    result &= newEmptyNode()

proc newVarDecl*(name: Node; typeExpr: Node; expr: Node = nil): Node =
    result = newVarDecl([name], typeExpr, expr)

proc newExprList*(exprs: openArray[Node]): Node =
    assert(exprs.allIt(it != nil))

    result  = newNode(nkExprList)
    result &= exprs

proc newExprList*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprList)
    result &= expr

proc newEmptyExprList*(): Node =
    result = newExprList([])
