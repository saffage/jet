type

  Typekind* = enum
    EMPTY

  TypeRef* = ref Type
  Type* = object
    id*   : string
    kind* : Typekind

  Symbolkind* = enum
    EMPTY

  SymbolRef* = ref Symbol
  Symbol* = object
    id*     : string
    kind*   : Symbolkind
    `type`* : Type
