## Very simple logger with a lot of hidden side effects.

import std/strutils

import ./line_info
import ./text_style


type LogLevel* = enum
    All
    Debug
    Info
    Hint
    Warn
    Error
    Panic
    None

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

const
    lineInfoFmt = "[$#] "
    logLevelFmt = "$#: "

template hasColors(): bool =
    true

func formatStackTrace(entries: seq[StackTraceEntry]): string =
    result = ""

    for entry in entries:
        let str = try: "at \"$#($#)\" in '$#'\n" % [
            $entry.filename,
            $entry.line,
            $entry.procname,
        ] except ValueError: "<fmt-error>\n"
        result.add(str)

func printStackTrace() =
    {.cast(noSideEffect).}:
        debugEcho(formatStackTrace(getStackTraceEntries()))

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
    const tag = logLevelFmt % $level
    print(level, tag, msg, colors)

func log(level: static[LogLevel]; msg: string; colors: bool; info: LineInfo) =
    const tag = logLevelFmt % $level
    print(level, (lineInfoFmt % $info) & tag, msg, colors)

func debug*(msg: string; info: LineInfo) =
    log(Debug, msg, hasColors(), info)

func debug*(msg: string) =
    log(Debug, msg, hasColors())

func info*(msg: string; info: LineInfo) =
    log(Info, msg, hasColors(), info)

func info*(msg: string) =
    log(Info, msg, hasColors())

func hint*(msg: string; info: LineInfo) =
    log(Hint, msg, hasColors(), info)

    {.cast(noSideEffect).}:
        inc(hints)

func hint*(msg: string) =
    log(Hint, msg, hasColors())

    {.cast(noSideEffect).}:
        inc(hints)

func warn*(msg: string; info: LineInfo) =
    log(Warn, msg, hasColors())

    {.cast(noSideEffect).}:
        inc(warns)

func warn*(msg: string) =
    log(Warn, msg, hasColors())

    {.cast(noSideEffect).}:
        inc(warns)

func error*(msg: string; info: LineInfo; errorCode=QuitFailure) =
    log(Error, msg, hasColors(), info)

    {.cast(noSideEffect).}:
        inc(errors)
        if printStackTraceOnError: printStackTrace()
        if errors >= maxErrors: quit(errorCode)

func error*(msg: string; errorCode=QuitFailure) =
    log(Error, msg, hasColors())

    {.cast(noSideEffect).}:
        inc(errors)
        if printStackTraceOnError: printStackTrace()
        if errors >= maxErrors: quit(errorCode)

func panic*(msg: string; info: LineInfo; errorCode=QuitFailure)
    {.noreturn.} =
    log(Panic, msg, hasColors(), info)

    {.cast(noSideEffect).}:
        if printStackTraceOnPanic: printStackTrace()
    quit(errorCode)

func panic*(msg: string; errorCode=QuitFailure)
    {.noreturn.} =
    log(Panic, msg, hasColors())

    {.cast(noSideEffect).}:
        if printStackTraceOnPanic: printStackTrace()
    quit(errorCode)
