import std/strformat
import std/sequtils
import std/math

import jet/ast/nodetypes
import jet/literal
import jet/token
import utils


export nodetypes


template `[]`*(self: Node; i: Natural): Node = self.children[i]
template `[]`*(self: Node; i: BackwardsIndex): Node = self.children[i]
template `[]`*[U; V](self: Node; i: HSlice[U, V]): seq[Node] = self.children[i]
template `[]=`*(self: Node; i: Natural; node: Node) = self.children[i] = node
template `[]=`*(self: Node; i: BackwardsIndex; node: Node) = self.children[i] = node
template `&=`*(self: Node; node: Node) = self.children.add(node)

template last*(self: Node): Node = self.children[^1]
template first*(self: Node): Node = self.children[1]
template `last=`*(self: Node; node: Node) = self.children[^1] = node
template `first=`*(self: Node; node: Node) = self.children[1] = node

template len*(self: Node): int = self.children.len()
template add*(self: Node; node: Node) = self.children.add(node)
template add*(self: Node; node: openArray[Node]) = self.children.add(node)
template insert*(self: Node; node: Node; i: Natural = 0) = self.children.insert(node, i)

proc expectKind*(self: Node; kind: NodeKind) =
    if self != nil and self.kind != kind:
        panic(fmt"expected {kind}, got {self.kind} instead")

proc expectKind*(self: Node; kinds: set[NodeKind]) =
    if self != nil and self.kind notin kinds:
        panic(fmt"expected one of {kinds}, got {self.kind} instead")

proc canHavePragma*(self: Node): bool =
    return self.kind in {nkLetStmt, nkDefStmt, nkTypedefStmt}

proc pragma*(self: Node): Node =
    result = case self.kind:
        of nkLetStmt: self[3]
        of nkDefStmt: self[4]
        of nkTypedefStmt: self[2]
        else: nil

proc `pragma=`*(self: Node; node: Node) =
    case self.kind:
        of nkLetStmt: self[3] = node
        of nkDefStmt: self[4] = node
        of nkTypedefStmt: self[2] = node
        else: panic(fmt"node of kind {self.kind} can't have a pragma")

proc newNode*(kind: NodeKind): Node =
    result = Node(kind: kind)

proc newEmptyNode*(): Node =
    result = Node(kind: nkEmpty)

proc newIdNode*(id: string): Node =
    result = Node(kind: nkId, id: id)

proc newLitNode*(lit: TypedLiteral): Node =
    result = Node(kind: nkLit, lit: lit)

proc newLitNode*(lit: Literal): Node =
    result = newLitNode(lit.toTypedLit())

proc id*(token: Token): Node =
    assert(token.kind in {TokenKind.Id, TokenKind.Underscore})
    newIdNode(token.value)

proc id*(identifier: string): Node =
    newIdNode(identifier)

proc traverseTree*(node: Node; buffer: var string; indentLevel: int = 0) =
    if indentLevel > 0:
        buffer.add('\n')

        for _ in 0 ..< indentLevel:
            buffer.add("|  ")

    if node == nil:
        buffer.add("null")
        return
    else:
        # Remove 'nk' prefix.
        buffer.add(($node.kind)[2 ..^ 1])

    case node.kind
    of nkEmpty:
        discard
    of nkLit:
        buffer.add($node.lit)
    of nkId:
        buffer.addQuoted(node.id)
    else:
        for n in node.children:
            n.traverseTree(buffer, indentLevel + 1)

proc treeRepr*(node: Node): string =
    result = ""
    node.traverseTree(result, 0)
