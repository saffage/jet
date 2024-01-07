import std/tables


type TypeKind* = enum
    tyUnknown = "<unknown>" ## Type needs to be inferred
    tyAny     = "<any>"     ## Type is any
    tyNil     = "nil"       ## Type of 'null' literal
    tyUnit    = "()"        ## Empty type
    tyNever   = "(!)"       ## Type of 'return' stmt
    tyISize   = "isize"     ## Pointer-sized signed integer
    tyUSize   = "usize"     ## Pointer-sized unsigned integer
    tyI8      = "i8"        ## Signed 8-bit integer
    tyI16     = "i16"       ## Signed 16-bit integer
    tyI32     = "i32"       ## Signed 32-bit integer
    tyI64     = "i64"       ## Signed 64-bit integer
    tyU8      = "u8"        ## Unsigned 8-bit integer
    tyU16     = "u16"       ## Unsigned 16-bit integer
    tyU32     = "u32"       ## Unsigned 32-bit integer
    tyU64     = "u64"       ## Unsigned 64-bit integer
    tyF32     = "f32"       ## Floating point 32-bit number
    tyF64     = "f64"       ## Floating point 64-bit number
    tyChar    = "char"      ## Character type
    tyBool    = "bool"      ## Boolean type
    tyString  = "string"    ## String
    tyFunc    = "func"      ## Function type
    tyStruct  = "struct"    ## Struct type
    tyEnum    = "enum"      ## Enumeration type

type TypeFlag* = enum
    EMPTY

const
    BuiltInTypes* = {tyNil .. tyBool}
    SizedTypes*   = {tyISize .. tyBool}

type Type* = ref object
    name*  : string
    flags* : set[TypeFlag]

    case kind* : TypeKind
    of tyUnknown, tyAny:
        nil
    of BuiltInTypes:
        discard
    of tyString:
        discard
    of tyFunc:
        funcParams*     : seq[Type]
        funcReturnType* : Type
        funcIsVarargs*  : bool
    of tyStruct:
        structFields* : OrderedTable[string, Type]
    of tyEnum:
        enumFields* : seq[string]

let unknownType* = Type(name: "", kind: tyUnknown)
let anyType*     = Type(name: "", kind: tyAny)
let nilType*     = Type(name: "nil", kind: tyNil)
let unitType*    = Type(name: "", kind: tyUnit)
let neverType*   = Type(name: "", kind: tyNever)
let isizeType*   = Type(name: "isize", kind: tyISize)
let usizeType*   = Type(name: "usize", kind: tyUSize)
let i8Type*      = Type(name: "i8", kind: tyI8)
let i16Type*     = Type(name: "i16", kind: tyI16)
let i32Type*     = Type(name: "i32", kind: tyI32)
let i64Type*     = Type(name: "i64", kind: tyI64)
let u8Type*      = Type(name: "u8", kind: tyU8)
let u16Type*     = Type(name: "u16", kind: tyU16)
let u32Type*     = Type(name: "u32", kind: tyU32)
let u64Type*     = Type(name: "u64", kind: tyU64)
let f32Type*     = Type(name: "f32", kind: tyF32)
let f64Type*     = Type(name: "f64", kind: tyF64)
let charType*    = Type(name: "char", kind: tyChar)
let boolType*    = Type(name: "bool", kind: tyBool)
let stringType*  = Type(name: "string", kind: tyString)

let builtInTypes* = [
    nilType,
    unitType,
    neverType,
    isizeType,
    usizeType,
    i8Type,
    i16Type,
    i32Type,
    i64Type,
    u8Type,
    u16Type,
    u32Type,
    u64Type,
    f32Type,
    f64Type,
    charType,
    boolType,
    stringType,
]

proc `$`*(self: Type): string =
    if self == nil:
        result = "<unknown>"
    else:
        if self.kind in {tyStruct, tyEnum}:
            result = $self.kind & " " & self.name
        else:
            result = $self.kind

func toTypeKind*(kind: string): TypeKind =
    case kind:
    of "nil"       : tyNil
    of "()"        : tyUnit
    of "(!)"       : tyNever
    of "isize"     : tyISize
    of "usize"     : tyUSize
    of "i8"        : tyI8
    of "i16"       : tyI16
    of "i32"       : tyI32
    of "i64"       : tyI64
    of "u8"        : tyU8
    of "u16"       : tyU16
    of "u32"       : tyU32
    of "u64"       : tyU64
    of "f32"       : tyF32
    of "f64"       : tyF64
    of "char"      : tyChar
    of "bool"      : tyBool
    of "string"    : tyString
    of "func"      : tyFunc
    of "struct"    : tyStruct
    else           : tyUnknown

proc getType*(kind: TypeKind): Type =
    result = case kind:
        of tyNever  : neverType
        of tyNil    : nilType
        of tyUnit   : unitType
        of tyISize  : isizeType
        of tyUSize  : usizeType
        of tyI8     : i8Type
        of tyI16    : i16Type
        of tyI32    : i32Type
        of tyI64    : i64Type
        of tyU8     : u8Type
        of tyU16    : u16Type
        of tyU32    : u32Type
        of tyU64    : u64Type
        of tyF32    : f32Type
        of tyF64    : f64Type
        of tyChar   : charType
        of tyBool   : boolType
        of tyString : stringType
        of tyFunc   : nil
        of tyStruct : nil
        else: nil

proc clone*(self: Type): Type =
    result = Type(name: self.name, kind: self.kind)

    case self.kind
    of tyFunc:
        result.funcParams     = self.funcParams
        result.funcReturnType = self.funcReturnType
        result.funcIsVarargs  = self.funcIsVarargs
    of tyStruct:
        result.structFields = self.structFields
    else:
        discard

proc newType*(name: string; kind: TypeKind): Type =
    result = Type(name: name, kind: kind)

proc newFuncType*(name: string; params: seq[Type]; returnType: Type; isVarargs: bool = false): Type =
    result = Type(
        name: name,
        kind: tyFunc,
        funcParams: params,
        funcReturnType: returnType,
        funcIsVarargs: isVarargs)

proc newStructType*(name: string; fields: OrderedTable[string, Type]): Type =
    result = Type(
        name: name,
        kind: tyStruct,
        structFields: fields)

proc newEnumType*(name: string; fields: seq[string]): Type =
    result = Type(
        name: name,
        kind: tyEnum,
        enumFields: fields)
