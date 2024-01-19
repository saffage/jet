import
  std/strformat,

  lib/utils


type
  TypeKind* = enum
    tyI32

  TypeRef* = ref Type
  Type* = object
    kind* : TypeKind

#
# Type
#

func `$`*(self: TypeRef): string =
  result =
    if self == nil:
      "nil"
    else:
      $self.kind

func sizeInBits*(self: TypeRef): int =
  result = case self.kind
    of tyI32: 32

func sizeInBytes*(self: TypeRef): int =
  result = case self.kind
    of tyI32: 4

type
  ScopeRef* = ref Scope

  Scope = object
    parent*  : ScopeRef
    symbols* : seq[SymbolRef]
    depth*   : int

  SymbolKind* = enum
    skType
    skVar
    skVal
    skFunc

  SymbolFlags* = enum
    EMPTY

  SymbolRef* = ref Symbol
  Symbol* = object
    id*     : string
    kind*   : SymbolKind
    `type`* : TypeRef
    scope*  : ScopeRef

#
# Scope
#

func newScope*(parent = nil.ScopeRef): ScopeRef =
  result = ScopeRef(parent: parent, symbols: @[])
  result.depth =
    if parent == nil: 0
    else: parent.depth + 1

func getSymbol*(self: ScopeRef; id: string): SymbolRef =
  var idx = self.symbols.findIt(it.id == id)
  result =
    if idx < 0: nil
    else: self.symbols[idx]

func getSymbolRec*(self: ScopeRef; id: string): SymbolRef =
  result = nil
  var scope = self
  while scope != nil:
    result = self.getSymbol(id)
    if result != nil: break
    scope = scope.parent

#
# Symbol
#

func `$`*(self: SymbolRef): string =
  result =
    if self == nil:
      "nil"
    else:
      &"{self.kind}: {self.id} {self.`type`}"
