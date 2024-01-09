import std/strformat

import jet/literal
import jet/ast/types
import jet/ast/sym


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
    nkFunc
        ## Id(name)
        ## Paren(params)
        ## expr(return-type)
        ## EqExpr(body)
        ## AnnotationList
    nkType
        ## Id(name)
        ## EqExpr(body)
        ## AnnotationList
    nkVar
    nkVal
    nkReturn
        ## expr
    nkIf
        ## IfBranch+
        ## ElseBranch?
    nkIfBranch
        ## expr(condition)
        ## DoExpr(body)
    nkElseBranch
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
    nkBracket
        ## expr(elements)*
    nkExprParen
        ## expr(prefix)
        ## Paren(elements)
    nkExprBrace
        ## expr(prefix)
        ## Brace(elements)
    nkExprBracket
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
    nkAnnotationList
        ## (Id | ExprParen)*
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
    nkExprList
        ## expr+

type NodeFlag* = enum
    EMPTY
    nfList
        ## For kinds `nkParen`, `nkBrace`, `nkBracket` means
        ## that the elements have been separated using `,`

type NodeFlags* = set[NodeFlag]

type Node* = ref object
    `type`* : Type
    sym*    : Sym
    flags*  : NodeFlags

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

    if self.flags != {}:
        result &= fmt" {$self.flags}"
