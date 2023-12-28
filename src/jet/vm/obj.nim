import std/tables
import std/strformat
import std/strutils
import std/hashes

import jet/ast
import jet/ast/sym
import jet/ast/types

import lib/utils


type Object* = ref object of Sym
    case oType* : TypeKind
    of tyISize  : isizeVal*  : int
    of tyUSize  : usizeVal*  : uint
    of tyI8     : i8Val*     : int8
    of tyI16    : i16Val*    : int16
    of tyI32    : i32Val*    : int32
    of tyI64    : i64Val*    : int64
    of tyU8     : u8Val*     : uint8
    of tyU16    : u16Val*    : uint16
    of tyU32    : u32Val*    : uint32
    of tyU64    : u64Val*    : uint64
    of tyF32    : f32Val*    : float32
    of tyF64    : f64Val*    : float64
    of tyChar   : charVal*   : char
    of tyBool   : boolVal*   : bool
    of tyString : stringVal* : string
    of tyEnum   : enumVal*   : int
    of tyFunc:
        fnParams* : seq[Sym]
        fnBody*   : Node
        fnScope*  : Scope
    of tyStruct:
        structFields* : OrderedTable[string, Object]
    of tyAny:
        # should not be instantiated
        nil
    of tyUnknown, tyNever, tyNull, tyUnit:
        nil


# ----- OBJECT ----- #
proc isHashableType*(kind: TypeKind): bool =
    result = kind in SizedTypes + {tyString}

proc hashKey*(self: Object): Hash =
    result = case self.oType:
        of tyISize  : hash(self.isizeVal)
        of tyUSize  : hash(self.usizeVal)
        of tyI8     : hash(self.i8Val)
        of tyI16    : hash(self.i16Val)
        of tyI32    : hash(self.i32Val)
        of tyI64    : hash(self.i64Val)
        of tyU8     : hash(self.u8Val)
        of tyU16    : hash(self.u16Val)
        of tyU32    : hash(self.u32Val)
        of tyU64    : hash(self.u64Val)
        of tyChar   : hash(self.charVal)
        of tyBool   : hash(self.boolVal)
        of tyString : hash(self.stringVal)
        else        : panic(fmt"type {self.oType} are not hashable")

proc copyVal*(self, other: Object) =
    ## Copies calue from `other` to `self`.
    assert(self != nil and other != nil)

    if self.oType != other.oType:
        panic(fmt"type of objects are not same ({self.oType} and {other.oType}), can't perform a copy of value")

    case self.oType
    of tyISize  : self.isizeVal = other.isizeVal
    of tyUSize  : self.usizeVal = other.usizeVal
    of tyI8     : self.i8Val = other.i8Val
    of tyI16    : self.i16Val = other.i16Val
    of tyI32    : self.i32Val = other.i32Val
    of tyI64    : self.i64Val = other.i64Val
    of tyU8     : self.u8Val = other.u8Val
    of tyU16    : self.u16Val = other.u16Val
    of tyU32    : self.u32Val = other.u32Val
    of tyU64    : self.u64Val = other.u64Val
    of tyChar   : self.charVal = other.charVal
    of tyBool   : self.boolVal = other.boolVal
    of tyString : self.stringVal = other.stringVal
    of tyFunc:
        self.fnBody   = nil
        self.fnParams = other.fnParams
        self.fnScope  = other.fnScope.clone()
    of tyStruct:
        self.structFields = other.structFields
    of tyEnum:
        self.enumVal = other.enumVal
    of tyUnknown, tyUnit, tyNull: discard
    else: unimplemented(fmt"'copyVal' for {self.oType}")

proc `$`*(self: Object): string =
    if self == nil:
        return "(null)"

    result = case self.oType:
        of tyNever  : "(!)"
        of tyUnit   : "()"
        of tyNull   : "null"
        of tyISize  : $self.isizeVal
        of tyUSize  : $self.usizeVal
        of tyI8     : $self.i8Val
        of tyI16    : $self.i16Val
        of tyI32    : $self.i32Val
        of tyI64    : $self.i64Val
        of tyU8     : $self.u8Val
        of tyU16    : $self.u16Val
        of tyU32    : $self.u32Val
        of tyU64    : $self.u64Val
        of tyF32    : $self.f32Val
        of tyF64    : $self.f64Val
        of tyBool   : $self.boolVal
        of tyChar   : $self.charVal
        of tyString : '"' & self.stringVal & '"'
        of tyEnum   : self.typ.name & '.' & self.typ.enumFields[self.enumVal]
        else        : unimplemented("'$' for Object")

proc inspect*(self: Object): string =
    if self == nil:
        return "null"

    case self.kind
    of skVar:
        result = fmt"var {self.id}"
    of skVal:
        result = fmt"val"
    of skLit, skReturnVal:
        result = "-"
    of skFunc:
        result =  fmt"fn {self.id}"
    of skType:
        result =  fmt"type {self.id}"
    else:
        unimplemented(fmt"inspect sym kind for {self.kind}")

    if self.oType == tyFunc:
        var params = newSeq[string]()
        for param in self.fnParams:
            params &= fmt"{param.typ.kind}"
        let fnParams     = params.join(", ")
        let fnReturnType = self.typ.funcReturnType
        result &= fmt" : func({fnParams}) {fnReturnType}"
    elif self.oType == tyStruct:
        result &= fmt" : {self.typ.name}"
    else:
        result &= fmt" : {self.oType}"

    case self.oType
    of tyUnit, tyNull, tyUnknown:
        discard
    of tyI8:
        result &= fmt" = {self.i8Val}"
    of tyI16:
        result &= fmt" = {self.i16Val}"
    of tyI32:
        result &= fmt" = {self.i32Val}"
    of tyI64:
        result &= fmt" = {self.i64Val}"
    of tyU8:
        result &= fmt" = {self.u8Val}"
    of tyU16:
        result &= fmt" = {self.u16Val}"
    of tyU32:
        result &= fmt" = {self.u32Val}"
    of tyU64:
        result &= fmt" = {self.u64Val}"
    of tyF32:
        result &= fmt" = {self.f32Val}"
    of tyF64:
        result &= fmt" = {self.f64Val}"
    of tyBool:
        result &= fmt" = {self.boolVal}"
    of tyString:
        result &= fmt" = ""{self.stringVal}"""
    # of oMap:
    #     var elems = newSeq[string]()
    #     for _, pair in self.mapVal.pairs():
    #         let keyStr = block:
    #             if pair.key.kind != oString: pair.key.str()
    #             else: '\"' & pair.key.str() & '\"'
    #         let valStr = block:
    #             if pair.val.kind != oString: pair.val.str()
    #             else:  '\"' & pair.val.str() & '\"'
    #         elems &= fmt"{keyStr}: {valStr}"
    #     let map = elems.join(", ")
    #     result &= fmt" = {{ {map} }}"
    of tyFunc:
        result &= " = {...}"
    # of tyReturnVal:
    #     result &= fmt" = val {self.retVal.inspect()}"
    # of tyType:
    of tyStruct:
        if self.typ.kind != tyStruct: unreachable()
        var fieldsStr = newSeq[string]()
        for fieldName, fieldValue in self.structFields.pairs():
            fieldsStr &= fmt"{fieldName} = {fieldValue}"
        let fields = fieldsStr.join(", ")
        result &= fmt(" = struct {self.typ.name} {{ {fields} }}")
    of tyEnum:
        result &= fmt(" = {self.typ.name}.{self.typ.enumFields[self.enumVal]}")
    else:
        unimplemented(fmt"inspect obj type for {self.oType}")
