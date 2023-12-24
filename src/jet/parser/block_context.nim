import std/strformat


type BlockContextKind* = enum
    Indent
    Column

type BlockContext* = object
    line* : int

    case kind* : BlockContextKind
    of Indent: indent* : int
    of Column: column* : int

proc initBlockContext*(kind: BlockContextKind; line: int; value: int): BlockContext =
    result = BlockContext(kind: kind, line: line)

    case kind:
        of Indent: result.indent = value
        of Column: result.column = value

func getColumn*(self: BlockContext): int =
    result = case self.kind:
        of Indent: self.indent
        of Column: self.column

func `==`*(self, other: BlockContext): bool =
    result = self.line == other.line and
             self.kind == other.kind and
             self.getColumn() == other.getColumn()

func `<`*(self, other: BlockContext): bool =
    result = self.kind == other.kind and
             self.getColumn() < other.getColumn()

func `<=`*(self, other: BlockContext): bool =
    result = self.kind == other.kind and
             self.getColumn() <= other.getColumn()

func `==`*(self: BlockContext; value: int): bool = self.getColumn() == value

func `<`*(self: BlockContext; value: int): bool = self.getColumn() < value

func `<=`*(self: BlockContext; value: int): bool = self.getColumn() <= value

func `==`*(value: int; self: BlockContext): bool = value == self.getColumn()

func `<`*(value: int; self: BlockContext): bool = value < self.getColumn()

func `<=`*(value: int; self: BlockContext): bool = value <= self.getColumn()

proc `$`*(self: BlockContext): string =
    result = fmt"BlockContext(line = {self.line}, {$self.kind} = {self.getColumn()})"
