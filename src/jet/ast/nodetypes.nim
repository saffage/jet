import std/strformat

import jet/literal


type NodeKind* = enum
    #[ Leaf Nodes ]#
    nkEmpty
        ## Empty node
    nkProgram
        ## File
    nkId
        ## Unchecked symbol
    nkGenericId
        ## Unchecked generic symbol
    nkLit
        ## Some typed literal

    #[ Other ]#
    nkDefStmt
        ## Id(name) | ExprDotExpr(instance-and-name)
        ## Paren(params)
        ## expr(return-type)
        ## EqExpr(body)
        ## PragmaList
    nkTypedefStmt
        ## Id(name)
        ## EqExpr(body)
        ## PragmaList
    nkReturnStmt
        ## expr
    nkIfExpr
        ## IfBranch+
        ## ElseBranch?
    nkIfBranch
        ## expr(condition)
        ## DoExpr(body)
    nkElseBranch
        ## stmt+
    nkDoExpr
        ## stmt+
    nkEqExpr
        ## stmt+
    nkExprEqExpr
        ## expr(left)
        ## expr(right)
    nkExprColonExpr
        ## expr(left)
        ## expr(right)
    nkExprDotExpr
        ## expr(left)
        ## expr(right)

    nkParen
        ## expr(elements)*
    nkBrace
        ## expr(elements)*
    nkExprParen
        ## expr(prefix)
        ## Paren(elements)
    nkExprBrace
        ## expr(prefix)
        ## Brace(elements)
    nkPrefix
        ## Id(op)
        ## expr(operand)
    nkPostfix
        ## Id(op)
        ## expr(operand)
    nkInfix
        ## Id(op)
        ## expr(left-operand)
        ## expr(right-operand)
    nkPragma
        ## Id(name)
        ## Paren(args)?
    nkPragmaList
        ## Pragma*
    nkVarDecl
        ## Id(name)+
        ## expr(type)
        ## EqExpr(value)
        ## PragmaList(pragmas)
    nkMatch
        ## expr
        ## (Case | IfBranch)+
        ## ElseBranch?
    nkCase
        ## expr(match-expression)
        ## DoExpr(body) | IfBranch(guard)
    nkVariant
        ## Infix | Id

type Node* = ref object
    case kind* : NodeKind
    of nkEmpty:
        nil
    of nkId, nkGenericId:
        id* : string
    of nkLit:
        lit* : TypedLiteral
    else:
        children* : seq[Node]

func isLeaf*(self: Node): bool =
    result = self.kind in {nkEmpty, nkId, nkLit}

func `$`*(self: Node): string =
    result = ($self.kind)[2..^1]

    case self.kind
    of nkLit:
        result &= $self.lit
    of nkId:
        result &= fmt(" \"{self.id}\"")
    else:
        discard
