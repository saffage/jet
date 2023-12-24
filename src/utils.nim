from std/math import log10, ceil

import std/macros
import utils/logger

export logger

## Most of bad ideas are located here.

func unreachable*(msg: string = "")
    {.noreturn, noinline, raises: [].} =
    if msg.len() > 0:
        panic("reached unreachable code: " & msg)
    else:
        panic("reached unreachable code")

func unimplemented*(msg: string = "")
    {.noreturn, noinline, raises: [].} =
    if msg.len() > 0:
        panic("unimplemented: " & msg)
    else:
        panic("unimplemented")

func todo*(msg: string = "")
    {.noreturn, noinline, raises: [].} =
    if msg.len() > 0:
        panic("todo: " & msg)
    else:
        panic("todo")

template unreachable*(cond: untyped; msg: string = "") =
    if not cond: unreachable(msg)

func numLen*(x: int): int =
    result = ord(x < 0) + log10(float32(abs(x)) + 1).ceil().int

func getPragmaIdent(pragma: NimNode): NimNode =
    return case pragma.kind:
        of nnkIdent:
            pragma
        of nnkExprColonExpr, nnkBracketExpr:
            pragma[0]
        else:
            nil

func hasPragma(fn: NimNode; pragma: string): bool =
    for pr in fn:
        let pragmaIdent = pr.getPragmaIdent()
        if pragmaIdent != nil and $pragmaIdent == pragma:
            return true
    return false

type AbstractMethodCallDefect = object of Defect

macro abstract*(fn: untyped): untyped =
    fn.expectKind(nnkMethodDef)

    if fn.body.kind != nnkEmpty:
        macros.error("an abstract method should not have an implementation", fn.body)
    if not fn.hasPragma("base"):
        fn.addPragma(ident"base")

    let name = $fn.name
    fn.body = quote do:
        raise newException(AbstractMethodCallDefect, "method '" & `name` & "' is abstract and should not be called")

    result = fn

template findIt*(collection: typed; pred: untyped): untyped =
    type Item = typeof( default( typeof(pairs(collection), typeOfIter) )[0] )
    var result = default(Item)
    for i, it {.inject.} in pairs(collection):
        if pred:
            result = i
            break
    result

template rfindIt*(collection: typed; pred: untyped): untyped =
    type Item = typeof( default( typeof(pairs(collection), typeOfIter) )[0] )
    var result = default(Item)
    for i, it {.inject.} in pairs(collection):
        if pred:
            result = i
            # this is so stupid but works
    result

template notNil*(x: untyped): untyped =
    var y = x
    if y.isNil():
        unreachable()
    else:
        y
