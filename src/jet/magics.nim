import
  std/enumutils,
  std/setutils

{.push, raises: [].}

type
  MagicKind* = enum
    ## Keep in sync with `modules.Module.magics`
    # mTypeIsize
    # mTypeUsize
    mTypeI8
    mTypeI16
    mTypeI32
    mTypeI64
    # mTypeU8
    # mTypeU16
    # mTypeU32
    # mTypeU64
    # mTypeF32
    # mTypeF64
    # mTypeChar
    # mTypeBool
    # mTypeAny
    # mTypeNil
    # mFuncPrint
    # mFuncPrintln
    # mFuncPanic

func `$`*(self: MagicKind): string =
  result = self.symbolName()[1 ..^ 1]

var
  resolvedMagics: set[MagicKind]

proc isResolved*(magic: MagicKind): bool =
  result = magic in resolvedMagics

proc getUnresolvedMagics*(): set[MagicKind] =
  result = resolvedMagics.complement()

proc getResolvedMagics*(): set[MagicKind] =
  result = resolvedMagics

proc markAsResolved*(magic: MagicKind)
  {.raises: [ValueError].} =
  if magic.isResolved():
    raise (ref ValueError)(msg: "magic '" & $magic & "' is already resolved")

  resolvedMagics.incl(magic)

{.pop.} # raises: []
