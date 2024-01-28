## Most of bad ideas are located here.

from std/math import log10, ceil

import
  std/macros,

  lib/logger

{.push, raises: [].}

type
  UnreachableDefect        = object of Defect
  UnimplementedDefect      = object of Defect
  TodoDefect               = object of Defect
  AbstractMethodCallDefect = object of Defect

func unreachable*(msg: string)
  {.noreturn, noinline.} =
  let msg = "reached unreachable code; " & msg
  logger.error(msg)
  raise newException(UnreachableDefect, msg)

func unreachable*()
  {.noreturn, noinline.} =
  const msg = "reached unreachable code"
  logger.error(msg)
  raise newException(UnreachableDefect, msg)

func unimplemented*(msg: string)
  {.noreturn, noinline.} =
  let msg = "unimplemented; " & msg
  logger.error(msg)
  raise newException(UnimplementedDefect, msg)

func unimplemented*()
  {.noreturn, noinline.} =
  const msg = "unimplemented"
  logger.error(msg)
  raise newException(UnimplementedDefect, msg)

func todo*(msg: string)
  {.noreturn, noinline.} =
  let msg = "TODO: " & msg
  logger.error(msg)
  raise newException(TodoDefect, msg)

func todo*()
  {.noreturn, noinline.} =
  const msg = "TODO"
  logger.error(msg)
  raise newException(TodoDefect, msg)

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
  var result = -1
  for i, it {.inject.} in pairs(collection):
    if pred:
      result = i
      break
  result

template rfindIt*(collection: typed; pred: untyped): untyped =
  type Item = typeof( default( typeof(pairs(collection), typeOfIter) )[0] )
  var result = -1
  for i, it {.inject.} in pairs(collection):
    if pred:
      result = i
      # this is so stupid but works
  result

template notNil*(x: typed) =
  var y = x
  if y.isNil():
    unreachable()
  else:
    y

{.pop.} # raises: []
