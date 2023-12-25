import std/tables

import jet/ast/types

import utils


proc isEqTypes*(x, y: Type): bool =
    if x == nil or y == nil or x.kind == tyUnknown or y.kind == tyUnknown:
        return false
    if x.kind == tyAny or y.kind == tyAny:
        return true
    if x.kind != y.kind:
        return false

    case x.kind
    of tyFunc:
        if not isEqTypes(x.funcReturnType, y.funcReturnType) or
                         x.funcIsVarargs != y.funcIsVarargs or
                         x.funcParams.len() != y.funcParams.len():
            return false
        for i in 0 ..< x.funcParams.len():
            if not isEqTypes(x.funcParams[i], y.funcParams[i]):
                return false
    of tyStruct:
        if x.structFields.len() != y.structFields.len():
            return false
        for fieldId in x.structFields.keys():
            if not(fieldId in y.structFields and
                   isEqTypes(x.structFields[fieldId], y.structFields[fieldId])):
                return false
    of tyEnum:
        return (x.name == y.name)
    of BuiltInTypes + {tyString}:
        discard
    of tyUnknown, tyAny:
        unreachable()

    return true

proc implicitConvertibleTo*(x, y: Type): bool =
    result = false

proc checkFuncTypes*(fn: Type; args: seq[Type]): int =
    ## **Returns:** index of invalid argument, or `-1` otherwise.
    ##
    ## **Note:** the index can be greater than `args.high`to indicate
    ## the case when the number of arguments is less than needed.
    let isVarargs   = fn.funcIsVarargs
    let fnParamsLen = fn.funcParams.len() - isVarargs.int
    let fnArgsLen   = args.len()

    if isVarargs and fnParamsLen < 0:
        # reachable when fn type is made manually
        unreachable("func with varargs must have at least 1 parameter")

    if fnArgsLen > fnParamsLen and not isVarargs:
        return fnParamsLen # index of redundant argument

    if fnArgsLen < fnParamsLen:
        return fnArgsLen # index of absent argument

    for i in 0 ..< fnParamsLen:
        if not isEqTypes(fn.funcParams[i], args[i]):
            return i

    if fnArgsLen > fnParamsLen and isVarargs:
        let varargsType = fn.funcParams[^1]

        for i in fnParamsLen ..< fnArgsLen:
            if not isEqTypes(varargsType, args[i]):
                return i

    return -1

when isMainModule:
    import std/strformat

    let builtInSymTypes = [
        # Type(
        #     kind          : tyFunc,
        #     funcIsVarargs : false,
        #     funcParams    : @[]),
        # Type(
        #     kind          : tyFunc,
        #     funcIsVarargs : false,
        #     funcParams    : @[i32Type]),
        Type(
            kind          : tyFunc,
            funcIsVarargs : true,
            funcParams    : @[anyType]),
        # Type(
        #     kind          : tyFunc,
        #     funcIsVarargs : true,
        #     funcParams    : @[i32Type]),
        # Type(
        #     kind          : tyFunc,
        #     funcIsVarargs : true,
        #     funcParams    : @[i32Type, anyType]),
        # Type(
        #     kind          : tyFunc,
        #     funcIsVarargs : true,
        #     funcParams    : @[i32Type, stringType]),
    ]

    for symType in builtinSymTypes:
        echo fmt"# -------------- varargs = {symType.funcIsVarargs}, params = {symType.funcParams}"

        let paramsSet = [
            @[],
            @[i32Type],
            @[i32Type, stringType],
            @[stringType],
            @[anyType],
        ]
        for params in paramsSet:
            echo fmt"{checkFuncTypes(symType, params)} -> {params}"
