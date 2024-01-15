import
  lib/utils

type
  LiteralKind* = enum
    lkString = "string"
    lkChar   = "char"
    lkInt    = "int"
    lkFloat  = "float"
    lkBool   = "bool"
    lkNil    = "nil"

  Literal* = object
    case kind*  : LiteralKind
    of lkString : stringVal* : string
    of lkChar   : charVal*   : char
    of lkInt    : intVal*    : BiggestInt
    of lkFloat  : floatVal*  : BiggestFloat
    of lkBool   : boolVal*   : bool
    of lkNil    : nil

func `$`*(self: Literal): string =
  return case self.kind:
    of lkString :  self.stringVal
    of lkChar   : $self.charVal
    of lkInt    : $self.intVal
    of lkFloat  : $self.floatVal
    of lkBool   : $self.boolVal
    of lkNil    : "nil"

func pretty*(self: Literal): string =
  return case self.kind:
    of lkString : '"' & self.stringVal & '"'
    of lkChar   : "'" & $self.charVal & "'"
    of lkInt    : $self.intVal
    of lkFloat  : $self.floatVal
    of lkBool   : $self.boolVal
    of lkNil    : "nil"

func len*(self: Literal): int =
  return case self.kind:
    of lkString : self.stringVal.len()
    of lkChar   : unimplemented("char literal 'len'")
    of lkInt    : ($self.intVal).len()
    of lkFloat  : ($self.floatVal).len()
    of lkBool   : ($self.boolVal).len()
    of lkNil    : 3

func newLit*(value: typeof(nil)): Literal =
  result = Literal(kind: lkNil)

func newLit*(value: sink string): Literal =
  result = Literal(kind: lkString, stringVal: ensureMove(value))

func newLit*(value: char): Literal =
  result = Literal(kind: lkChar, charVal: value)

func newLit*(value: SomeSignedInt): Literal =
  result = Literal(kind: lkInt, intVal: value.BiggestInt)

func newLit*(value: SomeFloat): Literal =
  result = Literal(kind: lkFloat, floatVal: value.BiggestFloat)

func newLit*(value: bool): Literal =
  result = Literal(kind: lkBool, boolVal: value)
