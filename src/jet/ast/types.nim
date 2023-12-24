import std/tables


type TypeKind* = enum
    tyUnknown = "<unknown>" ## Type needs to be inferred
    tyAny     = "<any>"     ## Type is any
    tyNever   = "<never>"   ## Type of 'return' stmt
    tyNull    = "null"      ## Type of 'null' literal
    tyUnit    = "unit"      ## Empty type
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
    tyChar    = "char"      ## Character type (alias for 'u8')
    tyBool    = "bool"      ## Boolean type
    tyString  = "string"    ## String
    tyFunc    = "func"      ## Function type
    tyStruct  = "struct"    ## Struct type
    tyEnum    = "enum"      ## Enumeration type

type TypeFlag* = enum
    EMPTY

const
    BuiltInTypes* = {tyNever..tyBool}
    SizedTypes*   = {tyISize..tyBool}

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
let neverType*   = Type(name: "", kind: tyNever)
let nullType*    = Type(name: "null", kind: tyNull)
let unitType*    = Type(name: "unit", kind: tyUnit)
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
    nullType,
    unitType,
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
    of "<never>"   : tyNever
    of "null"      : tyNull
    of "unit"      : tyUnit
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
        of tyNull   : nullType
        of tyUnit   : unitType
        of tyISize  : i_sizeType
        of tyUSize  : u_sizeType
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
        name         : name,
        kind         : tyStruct,
        structFields : fields)

proc newEnumType*(name: string; fields: seq[string]): Type =
    result = Type(
        name       : name,
        kind       : tyEnum,
        enumFields : fields)

#[
import std/macros

macro variant(head, body: untyped): untyped =
    echo head.treeRepr

    head.expectKind({nnkIdent, nnkBracketExpr})

    var objectGenericParams = newEmptyNode()
    let enumVariant   = newNimNode(nnkTypeDef)
    let objectVariant = newNimNode(nnkTypeDef)
    let variants      = newNimNode(nnkEnumTy)
    let objectBody    = newNimNode(nnkObjectTy)
    variants.add(newEmptyNode())
    objectBody.add(newEmptyNode())
    objectBody.add(newEmptyNode())

    let name =
        if head.kind == nnkIdent: head
        else:
            if head.len() > 1:
                objectGenericParams = newNimNode(nnkGenericParams)
                for genericParam in head[1..^1]:
                    genericParam.expectKind({nnkIdent, nnkExprColonExpr})
                    if genericParam.kind == nnkIdent:
                        objectGenericParams.add(newIdentDefs(genericParam, newEmptyNode()))
                    else:
                        objectGenericParams.add(newIdentDefs(genericParam[0], genericParam[1]))
            head[0]

    let enumVariantName = ident(name.strVal & "Variant")
    enumVariantName.copyLineInfo(name)

    enumVariant.add(enumVariantName)
    enumVariant.add(newEmptyNode()) # pragma
    enumVariant.add(variants)

    objectVariant.add(name)
    objectVariant.add(objectGenericParams)
    objectVariant.add(objectBody)

    let caseIdent = ident"variant"
    let cases     = newNimNode(nnkRecCase)
    cases.add(newIdentDefs(caseIdent, enumVariant[0]))

    for variant in body:
        variant.expectKind({nnkIdent, nnkObjConstr, nnkCall})

        case variant.kind:
        of nnkIdent:
            let variantIdent = ident(variant.strVal)
            variants.add(variantIdent)

            let ofBranch = newNimNode(nnkOfBranch)
            ofBranch.add(variantIdent)
            ofBranch.add(newNimNode(nnkNilLit))

            cases.add(ofBranch)
        of nnkObjConstr:
            let variantIdent = ident(variant[0].strVal)
            variants.add(variantIdent)

            let fields = newNimNode(nnkRecList)
            for field in variant[1..^1]:
                fields.add(newIdentDefs(field[0], field[1]))

            let ofBranch = newNimNode(nnkOfBranch)
            ofBranch.add(variantIdent)
            ofBranch.add(fields)

            cases.add(ofBranch)
        of nnkCall:
            let variantIdent = ident(variant[0].strVal)
            variants.add(variantIdent)

            let fields = newNimNode(nnkRecList)
            for i, field in variant[1..^1]:
                fields.add(newIdentDefs(ident("i" & $i & variant[0].strVal), field))

            let ofBranch = newNimNode(nnkOfBranch)
            ofBranch.add(variantIdent)
            ofBranch.add(fields)

            cases.add(ofBranch)
        else:
            discard

    let recordList = newNimNode(nnkRecList)
    recordList.add(cases)
    objectBody.add(recordList)

    let typeSection = newNimNode(nnkTypeSection)
    typeSection.add(enumVariant)
    typeSection.add(objectVariant)

    result = newStmtList(typeSection)
 ]#
#[
StmtList
    TypeSection
        TypeDef
            Ident "OptionVariant"
            Empty
            EnumTy
                Empty
                Ident "SomeNamed"
                Ident "Some"
                Ident "None"
        TypeDef
            Ident "Option"
            GenericParams
                IdentDefs
                Ident "T"
                Empty
                Empty
            ObjectTy
                Empty
                Empty
                RecList
                    RecCase
                        IdentDefs
                            Ident "variant"
                            Ident "OptionVariant"
                            Empty
                        OfBranch
                            Ident "SomeNamed"
                            RecList
                                IdentDefs
                                    Ident "data"
                                    Ident "T"
                                    Empty
                        OfBranch
                            Ident "Some"
                            RecList
                                IdentDefs
                                    Ident "i0Some"
                                    Ident "T"
                                    Empty
                            OfBranch
                                Ident "None"
                                RecList
                                    NilLit
]#

# dumpTree:
#     type OptionVariant = enum
#         SomeNamed
#         Some
#         None
# dumpTree:
#     type Option[T] = object
#         case variant: OptionVariant
#         of SomeNamed:
#             data: T
#         of Some:
#             i0Some: T
#         of None:
#             nil

# variant Option[T: SomeInteger]:
#     SomeNamed(data: T)
#     Some(T)
#     None
