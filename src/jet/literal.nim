import std/strutils except escape
import std/strformat

from jet/scanner import escape

import lib/utils


# ----- UNTYPED LITERAL ----- #
type LiteralKind* = enum
    lkEmpty  = "empty"
    lkString = "string"
    lkChar   = "char"
    lkInt    = "int"
    lkUInt   = "uint"
    lkFloat  = "float"
    lkBool   = "bool"
    lkNull   = "null"

type Literal* = object
    case kind*  : LiteralKind
    of lkString : stringVal* : string
    of lkChar   : charVal*   : char
    of lkInt    : intVal*    : BiggestInt
    of lkUInt   : uintVal*   : BiggestUInt
    of lkFloat  : floatVal*  : BiggestFloat
    of lkBool   : boolVal*   : bool
    of lkNull   : nil
    of lkEmpty  : nil

func `$`*(self: Literal): string =
    return case self.kind:
        of lkString :  self.stringVal
        of lkChar   : $self.charVal
        of lkInt    : $self.intVal
        of lkUInt   : $self.uintVal
        of lkFloat  : $self.floatVal
        of lkBool   : $self.boolVal
        of lkNull   : "<null>"
        of lkEmpty  : "<empty>"

func len*(self: Literal): int =
    return case self.kind:
        of lkString : self.stringVal.len()
        of lkChar   : unimplemented("char literal 'len'")
        of lkInt    : ($self.intVal).len()
        of lkUInt   : ($self.uintVal).len()
        of lkFloat  : ($self.floatVal).len()
        of lkBool   : ($self.boolVal).len()
        of lkNull   : 4
        of lkEmpty  : -1

func newEmptyLiteral*(): Literal =
    result = Literal(kind: lkEmpty)

func newNullLiteral*(): Literal =
    result = Literal(kind: lkNull)

func newLit*(value: sink string): Literal =
    result = Literal(kind: lkString, stringVal: ensureMove(value))

func newLit*(value: char): Literal =
    result = Literal(kind: lkChar, charVal: value)

func newLit*(value: SomeSignedInt): Literal =
    result = Literal(kind: lkInt, intVal: value.BiggestInt)

func newLit*(value: SomeUnsignedInt): Literal =
    result = Literal(kind: lkUInt, uintVal: value.BiggestUInt)

func newLit*(value: SomeFloat): Literal =
    result = Literal(kind: lkFloat, floatVal: value.BiggestFloat)

func newLit*(value: bool): Literal =
    result = Literal(kind: lkBool, boolVal: value)


# ----- TYPED LITERAL ----- #
type TypedLiteralKind* = enum
    tlkNever  = "never"
    tlkNull   = "null"
    tlkUnit   = "unit"
    tlkISize  = "isize"
    tlkUSize  = "usize"
    tlkI8     = "i8"
    tlkI16    = "i16"
    tlkI32    = "i32"
    tlkI64    = "i64"
    tlkU8     = "u8"
    tlkU16    = "u16"
    tlkU32    = "u32"
    tlkU64    = "u64"
    tlkF32    = "f32"
    tlkF64    = "f64"
    tlkBool   = "bool"
    tlkChar   = "char"
    tlkString = "string"

type TypedLiteral* = object
    case kind*   : TypedLiteralKind
    of tlkString : stringVal* : string
    of tlkISize  : isizeVal*  : int
    of tlkUSize  : usizeVal*  : uint
    of tlkI8     : i8Val*     : int8
    of tlkI16    : i16Val*    : int16
    of tlkI32    : i32Val*    : int32
    of tlkI64    : i64Val*    : int64
    of tlkU8     : u8Val*     : uint8
    of tlkU16    : u16Val*    : uint16
    of tlkU32    : u32Val*    : uint32
    of tlkU64    : u64Val*    : uint64
    of tlkF32    : f32Val*    : float32
    of tlkF64    : f64Val*    : float64
    of tlkBool   : boolVal*   : bool
    of tlkChar   : charVal*   : char
    of tlkNull   : nil
    of tlkUnit   : nil
    of tlkNever  : nil

func newTypedNeverLit*(): TypedLiteral = TypedLiteral(kind: tlkNever)
func newTypedNullLit*(): TypedLiteral = TypedLiteral(kind: tlkNull)
func newTypedUnitLit*(): TypedLiteral = TypedLiteral(kind: tlkUnit)

func newTypedLit*(value: sink string): TypedLiteral = TypedLiteral(kind: tlkString, stringVal: ensureMove(value))
func newTypedLit*(value: int): TypedLiteral = TypedLiteral(kind: tlkISize, isizeVal: value)
func newTypedLit*(value: uint): TypedLiteral = TypedLiteral(kind: tlkUSize, usizeVal: value)
func newTypedLit*(value: int8): TypedLiteral = TypedLiteral(kind: tlkI8, i8Val: value)
func newTypedLit*(value: int16): TypedLiteral = TypedLiteral(kind: tlkI16, i16Val: value)
func newTypedLit*(value: int32): TypedLiteral = TypedLiteral(kind: tlkI32, i32Val: value)
func newTypedLit*(value: int64): TypedLiteral = TypedLiteral(kind: tlkI64, i64Val: value)
func newTypedLit*(value: uint8): TypedLiteral = TypedLiteral(kind: tlkU8, u8Val: value)
func newTypedLit*(value: uint16): TypedLiteral = TypedLiteral(kind: tlkU16, u16Val: value)
func newTypedLit*(value: uint32): TypedLiteral = TypedLiteral(kind: tlkU32, u32Val: value)
func newTypedLit*(value: uint64): TypedLiteral = TypedLiteral(kind: tlkU64, u64Val: value)
func newTypedLit*(value: float32): TypedLiteral = TypedLiteral(kind: tlkF32, f32Val: value)
func newTypedLit*(value: float64): TypedLiteral = TypedLiteral(kind: tlkF64, f64Val: value)
func newTypedLit*(value: bool): TypedLiteral = TypedLiteral(kind: tlkBool, boolVal: value)
func newTypedLit*(value: char): TypedLiteral = TypedLiteral(kind: tlkChar, charVal: value)

func str*(self: TypedLiteral): string =
    ## **Returns:** a string representation of literal in a **Jet** code
    result = case self.kind:
        of tlkISize  : $self.isizeVal & "i"
        of tlkUSize  : $self.usizeVal & "u"
        of tlkI8     : $self.i8Val & "i8"
        of tlkI16    : $self.i16Val & "i16"
        of tlkI32    : $self.i32Val
        of tlkI64    : $self.i64Val & "i64"
        of tlkU8     : $self.u8Val & "u8"
        of tlkU16    : $self.u16Val & "u16"
        of tlkU32    : $self.u32Val & "u32"
        of tlkU64    : $self.u64Val & "u64"
        of tlkF32    : $self.f32Val & "f32"
        of tlkF64    : $self.f64Val
        of tlkBool   : $self.boolVal
        of tlkChar   : "'" & escape($self.charVal) & "'"
        of tlkString : '"' & escape(self.stringVal) & '"'
        of tlkNull   : "null"
        of tlkUnit   : "()"
        of tlkNever  : "TODO \"never type expr\""

func `$`*(self: TypedLiteral): string =
    template escape(str: string): string =
        multiReplace(str, ("\0", "\\0"), ("\n", "\\n"), ("\r", "\\r"), ("\t", "\\t"), ("\'", "\\'"), ("\"", "\\\""), ("\\", "\\\\"))

    result = case self.kind:
        of tlkISize  : $self.isizeVal
        of tlkUSize  : $self.usizeVal
        of tlkI8     : $self.i8Val
        of tlkI16    : $self.i16Val
        of tlkI32    : $self.i32Val
        of tlkI64    : $self.i64Val
        of tlkU8     : $self.u8Val
        of tlkU16    : $self.u16Val
        of tlkU32    : $self.u32Val
        of tlkU64    : $self.u64Val
        of tlkF32    : $self.f32Val
        of tlkF64    : $self.f64Val
        of tlkBool   : $self.boolVal
        of tlkChar   : "'" & escape($self.charVal) & "'"
        of tlkString : '"' & escape(self.stringVal) & '"'
        of tlkNull   : "<null>"
        of tlkUnit   : "<unit>"
        of tlkNever  : "<never>"
    result = fmt"(type: {self.kind}, value: {result})"

func intRangeCheck*[T, R](value: T; range: Slice[R]): bool =
    when T is SomeSignedInt:
        when R isnot SomeSignedInt: {.error: "R must be SomeSignedInt".}
    elif T is SomeUnsignedInt:
        when R isnot SomeUnsignedInt: {.error: "R must be SomeSignedInt".}
    else: {.error: "T must be SomeInteger".}

    # convert params to bigger type
    when sizeof(T) < sizeof(R):
        let rangeA = range.a
        let rangeB = range.b
        let value  = R(value)
    elif sizeof(T) > sizeof(R):
        let rangeA = T(range.a)
        let rangeB = T(range.b)
    else:
        let rangeA = range.a
        let rangeB = range.b

    result = (value >= rangeA and value <= rangeB)

func toTypedLit*(lit: Literal): TypedLiteral =
    result = case lit.kind:
        of lkInt:
            if intRangeCheck(lit.intVal, int32.low .. int32.high):
                newTypedLit(lit.intVal.int32)
            else:
                newTypedLit(lit.intVal.int64)
        of lkUInt:
            if intRangeCheck(lit.uintVal, uint32.low .. uint32.high):
                newTypedLit(lit.uintVal.uint32)
            else:
                newTypedLit(lit.uintVal.uint64)
        of lkFloat:
            newTypedLit(lit.floatVal.float64)
        of lkBool:
            newTypedLit(lit.boolVal)
        of lkNull:
            newTypedNullLit()
        of lkEmpty:
            newTypedUnitLit()
        of lkString:
            newTypedLit(lit.stringVal)
        of lkChar:
            newTypedLit(lit.charVal)

func tryIntoTyped*(lit: Literal; kind: TypedLiteralKind): TypedLiteral =
    template checkKind(expectedKinds: set[TypedLiteralKind]): untyped =
        if kind notin expectedKinds:
            panic(fmt"invalid literal kind for typed literal, got {lit.kind} for {kind}")

    template checkKind(expectedKind: TypedLiteralKind): untyped =
        if kind != expectedKind:
            panic(fmt"invalid literal kind for typed literal, got {lit.kind} for {kind}")

    template checkRangeInt(range: Slice): untyped =
        if not intRangeCheck(lit.intVal, range):
            panic(fmt"value '{lit.intVal}' not in range {$range}")

    template checkRangeUInt(range: Slice): untyped =
        if not intRangeCheck(lit.uintVal, range):
            panic(fmt"value '{lit.uintVal}' not in range {$range}")

    result = case lit.kind:
        of lkInt:
            checkKind({tlkISize, tlkI8, tlkI16, tlkI32, tlkI64})

            case kind
            of tlkISize:
                checkRangeInt(int.low.int64 .. int.high.int64)
                newTypedLit(lit.intVal.int)
            of tlkI8:
                checkRangeInt(int8.low.int64 .. int8.high.int64)
                newTypedLit(lit.intVal.int8)
            of tlkI16:
                checkRangeInt(int16.low.int64 .. int16.high.int64)
                newTypedLit(lit.intVal.int16)
            of tlkI32:
                checkRangeInt(int32.low.int64 .. int32.high.int64)
                newTypedLit(lit.intVal.int32)
            of tlkI64:
                checkRangeInt(int64.low .. int64.high)
                newTypedLit(lit.intVal.int64)
            else: unreachable()
        of lkUInt:
            checkKind({tlkUSize, tlkU8, tlkU16, tlkU32, tlkU64})

            case kind
            of tlkUSize:
                checkRangeUInt(uint.low.uint64 .. uint.high.uint64)
                newTypedLit(lit.uintVal.uint)
            of tlkU8:
                checkRangeUInt(uint8.low.uint64 .. uint8.high.uint64)
                newTypedLit(lit.uintVal.uint8)
            of tlkU16:
                checkRangeUInt(uint16.low.uint64 .. uint16.high.uint64)
                newTypedLit(lit.uintVal.uint16)
            of tlkU32:
                checkRangeUInt(uint32.low.uint64 .. uint32.high.uint64)
                newTypedLit(lit.uintVal.uint32)
            of tlkU64:
                checkRangeUInt(uint64.low .. uint64.high)
                newTypedLit(lit.uintVal.uint64)
            else: unreachable()
        of lkFloat:
            checkKind({tlkF32, tlkF64})
            newTypedLit(lit.floatVal.float64)
        of lkBool:
            checkKind(tlkBool)
            newTypedLit(lit.boolVal)
        of lkNull:
            checkKind(tlkNull)
            newTypedNullLit()
        of lkEmpty:
            checkKind(tlkUnit)
            newTypedUnitLit()
        of lkString:
            checkKind(tlkString)
            newTypedLit(lit.stringVal)
        of lkChar:
            checkKind(tlkChar)
            newTypedLit(lit.charVal)


when isMainModule:
    import std/unittest

    suite "Literals":
        template intRangeCheckTest(T: typedesc) {.dirty.} =
            check(intRangeCheck(T(0), T.low       .. T.high))
            check(intRangeCheck(T(0), T.low.int64 .. T.high.int64))

            check(intRangeCheck(T(-1), T.low       .. T.high))
            check(intRangeCheck(T(-1), T.low.int64 .. T.high.int64))

            check(intRangeCheck(T.low, T.low        .. T.high))
            check(intRangeCheck(T.low, T.low.int64  .. T.high.int64))
            check(intRangeCheck(T.high, T.low       .. T.high))
            check(intRangeCheck(T.high, T.low.int64 .. T.high.int64))

            when sizeof(T) < 8:
                check(not intRangeCheck(T.low.int64.pred, T.low        .. T.high))
                check(not intRangeCheck(T.low.int64.pred, T.low.int64  .. T.high.int64))
                check(not intRangeCheck(T.high.int64.succ, T.low       .. T.high))
                check(not intRangeCheck(T.high.int64.succ, T.low.int64 .. T.high.int64))

        template uintRangeCheckTest(T: typedesc) {.dirty.} =
            check(intRangeCheck(T(0), T.low       .. T.high))
            check(intRangeCheck(T(0), T.low.uint64 .. T.high.uint64))

            check(intRangeCheck(T.low, T.low        .. T.high))
            check(intRangeCheck(T.low, T.low.uint64  .. T.high.uint64))
            check(intRangeCheck(T.high, T.low       .. T.high))
            check(intRangeCheck(T.high, T.low.uint64 .. T.high.uint64))

            when sizeof(T) < 8:
                check(not intRangeCheck(T.low.uint64.pred, T.low        .. T.high))
                check(not intRangeCheck(T.low.uint64.pred, T.low.uint64  .. T.high.uint64))
                check(not intRangeCheck(T.high.uint64.succ, T.low       .. T.high))
                check(not intRangeCheck(T.high.uint64.succ, T.low.uint64 .. T.high.uint64))

        test "func 'intRangeCheck'":
            intRangeCheckTest(int8)
            intRangeCheckTest(int16)
            intRangeCheckTest(int32)
            intRangeCheckTest(int64)
            uintRangeCheckTest(uint8)
            uintRangeCheckTest(uint16)
            uintRangeCheckTest(uint32)
            uintRangeCheckTest(uint64)

        test "func 'toTypedLit'":
            let lit1 = newLit(int32.high.int64.pred)
            let lit2 = newLit(int32.high.int64.succ)
            check(lit1.kind == lkInt)
            check(lit2.kind == lkInt)

            let typedLit1 = toTypedLit(lit1)
            let typedLit2 = toTypedLit(lit2)
            check(typedLit1.kind == tlkI32)
            check(typedLit2.kind == tlkI64)

            let lit3 = newLit(uint32.high.uint64.pred)
            let lit4 = newLit(uint32.high.uint64.succ)
            check(lit3.kind == lkUInt)
            check(lit4.kind == lkUInt)

            let typedLit3 = toTypedLit(lit3)
            let typedLit4 = toTypedLit(lit4)
            check(typedLit3.kind == tlkU32)
            check(typedLit4.kind == tlkU64)
