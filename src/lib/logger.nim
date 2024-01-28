## Very simple logger with a lot of hidden side effects.

import
  std/strutils,
  std/strformat,

  ./lineinfo,
  ./colors

{.push, raises: [].}

type
  LogLevel* = enum
    All
    Debug
    Info
    Hint
    Warn
    Error
    Panic
    None

  LoggerDefect* = object of Defect

const
  debugTagStyle* = TextStyle(foreground: Magenta, bold: true)
  debugMsgStyle* = TextStyle(foreground: White)
  hintTagStyle*  = TextStyle(foreground: Cyan)
  hintMsgStyle*  = TextStyle(foreground: White)
  warnTagStyle*  = TextStyle(foreground: Yellow)
  warnMsgStyle*  = TextStyle(foreground: White)
  errorTagStyle* = TextStyle(foreground: BrightRed, bold: true)
  errorMsgStyle* = TextStyle(foreground: White)
  panicTagStyle* = TextStyle(foreground: Red, bold: true)
  panicMsgStyle* = TextStyle(foreground: Red)

func tagStyle(self: LogLevel): TextStyle =
  result = case self:
    of Debug : debugTagStyle
    of Hint  : hintTagStyle
    of Warn  : warnTagStyle
    of Error : errorTagStyle
    of Panic : panicTagStyle
    else: TextStyle()

func msgStyle(self: LogLevel): TextStyle =
  result = case self:
    of Debug : debugMsgStyle
    of Hint  : hintMsgStyle
    of Warn  : warnMsgStyle
    of Error : errorMsgStyle
    of Panic : panicMsgStyle
    else: TextStyle()

var
  loggingLevel* =
    when defined(release): LogLevel.Error
    else: LogLevel.All
  printStackTraceOnError* = false
  printStackTraceOnPanic* = true

  hints*     = 0
  warns*     = 0
  errors*    = 0
  maxErrors* = 1

template hasColors(): bool =
  true

func print(level: static[LogLevel]; tag, msg: string; colors: bool) =
  {.cast(noSideEffect).}:
    if level < loggingLevel: return

  if colors:
    let tag = tag @ level.tagStyle()
    let msg = msg @ level.msgStyle()
    debugEcho(tag & msg)
  else:
    debugEcho(tag & msg)

func log(level: static[LogLevel]; msg: string; colors: bool) =
  const tag = toLowerAscii($level) & ": "
  print(level, tag, msg, colors)

func log(level: static[LogLevel]; msg: string; colors: bool; pos: FilePosition) =
  const levelTag = toLowerAscii($level)
  let tag = try: &"{levelTag}[{pos}]: " except ValueError: "<fmt-error>"
  print(level, tag, msg, colors)

func debug*(msg: string; pos: FilePosition) =
  log(Debug, msg, hasColors(), pos)

func debug*(msg: string) =
  log(Debug, msg, hasColors())

func pos*(msg: string; pos: FilePosition) =
  log(Info, msg, hasColors(), pos)

func pos*(msg: string) =
  log(Info, msg, hasColors())

func hint*(msg: string; pos: FilePosition) =
  log(Hint, msg, hasColors(), pos)

  {.cast(noSideEffect).}:
    inc(hints)

func hint*(msg: string) =
  log(Hint, msg, hasColors())

  {.cast(noSideEffect).}:
    inc(hints)

func warn*(msg: string; pos: FilePosition) =
  log(Warn, msg, hasColors())

  {.cast(noSideEffect).}:
    inc(warns)

func warn*(msg: string) =
  log(Warn, msg, hasColors())

  {.cast(noSideEffect).}:
    inc(warns)

func error*(msg: string; pos: FilePosition) =
  log(Error, msg, hasColors(), pos)

  {.cast(noSideEffect).}:
    inc(errors)
    if errors >= maxErrors:
      raise newException(LoggerDefect, msg)

func error*(msg: string) =
  log(Error, msg, hasColors())

  {.cast(noSideEffect).}:
    inc(errors)
    if errors >= maxErrors:
      raise newException(LoggerDefect, msg)

func panic*(msg: string; pos: FilePosition)
  {.noreturn.} =
  log(Panic, msg, hasColors(), pos)
  raise newException(LoggerDefect, msg)

func panic*(msg: string)
  {.noreturn.} =
  log(Panic, msg, hasColors())
  raise newException(LoggerDefect, msg)

{.pop.} # raises: []
