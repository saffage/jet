import
  std/strformat,
  std/enumutils,

  lib/utils

type
  TypeKind* = enum
    tyI8
    tyI16
    tyI32
    tyI64
    tyRef
    tyNil

  TypeRef* = ref Type
  Type* = object
    kind*   : TypeKind
    parent* : TypeRef = nil

#
# Type
#

func `$`*(self: TypeRef): string =
  result =
    if self == nil:
      "nil"
    elif self.parent != nil:
      $self.kind & " " & $self.parent
    else:
      $self.kind

func sizeInBits*(self: TypeRef): int =
  result = case self.kind
    of tyI8: 8
    of tyI16: 16
    of tyI32: 32
    of tyI64: 64
    of tyRef: sizeof(pointer) * 8
    else: 0

func sizeInBytes*(self: TypeRef): int =
  result = case self.kind
    of tyI8: 1
    of tyI16: 2
    of tyI32: 4
    of tyI64: 8
    of tyRef: sizeof(pointer)
    else: 0

func isCompatibleTypes*(self, other: TypeRef): bool =
  if self == nil or other == nil:
    return false

  result = case self.kind:
    of tyI8, tyI16, tyI32, tyI64:
      other.kind == self.kind
    of tyRef:
      other.kind == tyNil or (
        other.kind == tyRef and
        other.parent.kind == self.parent.kind
      )
    else:
      false

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

func `$`*(self: SymbolKind): string =
  result = self.symbolName()[2 ..^ 1]

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
