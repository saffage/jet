import
  std/strutils,

  pkg/questionable,

  ./enums,
  ./lineinfo,
  ./colors

{.push, raises: [].}

const
  messageStyle   = TextStyle(bold: true)
  noteStyle      = TextStyle(bold: true, foreground: BrightBlue)
  hintStyle      = TextStyle(bold: true, foreground: Cyan)
  warningStyle   = TextStyle(bold: true, foreground: Yellow)
  errorStyle     = TextStyle(bold: true, foreground: Red)
  lineNumStyle   = TextStyle(bold: true, foreground: BrightGreen)
  highlightStyle = TextStyle(underlined: true, foreground: BrightMagenta)

type
  HighlightInfoKind* {.pure.} = enum
    Hint
    Warning
    Error

  HighlightTarget* = object
    range*   : FileRange
    line*    : string
    details* : string

  HighlightNote* = object
    message*  : string
    filePath* : ?string
    range*    : ?FileRange

  HighlightInfo* = object
    message*  : string
    filePath* : string
    target*   : ?HighlightTarget
    notes*    : seq[HighlightNote]

    case kind* : HighlightInfoKind
    of Hint:
      hint* : CompilationHint
    of Warning:
      warning* : CompilationWarning
    of Error:
      error* : CompilationError

func `$`(kind: HighlightInfoKind): string =
  result = case kind:
    of Hint:
      "hint"
    of Warning:
      "warning"
    of Error:
      "error"

func code(info: HighlightInfo): int =
  result = case info.kind:
    of Hint: info.hint.int
    of Warning: info.warning.int
    of Error: info.error.int

func hasCode(info: HighlightInfo): bool =
  result = info.code() != 0

func codeStr(info: HighlightInfo): string =
  result = case info.kind:
    of Hint: $info.hint
    of Warning: $info.warning
    of Error: $info.error

func style(info: HighlightInfo): TextStyle =
  result = case info.kind:
    of Hint: hintStyle
    of Warning: warningStyle
    of Error: errorStyle

func appendFilePath(
  buf: var string;
  filePath: string;
  linePrefix: string;
  pos: ?FilePos = none(FilePos);
) =
  buf &= linePrefix
  buf &= stylizeText(" --> ", lineNumStyle)
  buf &= filePath

  if pos =? pos:
    buf &= ":"
    buf &= $pos

  buf &= "\n"

func visualize(info: HighlightInfo): string =
  const
    underscoreCharSet = ['~', '-', '^']
    lineNumSuffix     = " |"
    lineNumNoteSuffix = " ="
    filePathPrefix    = " --> "

  let
    lineNumStr =
      if target =? info.target:
        $target.range.a.line
      else:
        ""
    lineNumLen = lineNumStr.len()

  if info.filePath != "":
    let linePrefix = spaces(max(0, lineNumLen - 1))
    result.appendFilePath(info.filePath, linePrefix, info.target.?range.?a)

  let
    emptyLineNum = spaces(lineNumLen) & stylizeText(lineNumSuffix, lineNumStyle)

  if target =? info.target:
    let highlightRange =
      if target.range.b.line > target.range.a.line:
        target.range.a.column.pred.Natural .. target.line.len().Natural
      else:
        target.range.a.column.pred.Natural .. target.range.b.column.pred.Natural

    let
      lineNum        = stylizeText(lineNumStr & lineNumSuffix, lineNumStyle)
      line           = stylizeText(target.line, highlightStyle, highlightRange)
      underliningStr = repeat(underscoreCharSet[0], highlightRange.len())
      underlining    = stylizeText(underliningStr, info.style())

    result &= emptyLineNum & "\n"
    result &= lineNum & line & "\n"

    let
      emptyLineUntilHighlight = emptyLineNum & spaces(target.range.a.column.int - 1)

    result &= emptyLineUntilHighlight & underlining & '\n'

    if target.details != "":
      result &= emptyLineUntilHighlight & stylizeText(target.details, info.style()) & "\n"

    result &= emptyLineNum & "\n"

  let
    emptyNote = spaces(lineNumLen) & stylizeText(lineNumNoteSuffix, lineNumStyle)
    noteLabel = stylizeText(" note: ", noteStyle)

  for note in info.notes:
    result &= emptyNote & noteLabel & stylizeText(note.message, messageStyle) & "\n"

    if filePath =? note.filePath:
      result.appendFilePath(filePath, emptyLineNum & " ", note.range.?a)

    result &= emptyLineNum & "\n"

proc highlightInfoInFile*(info: HighlightInfo)
  {.raises: [IOError].} =
  let label = $info.kind & (
    if info.hasCode():
      "[" & $info.codeStr() & "]:"
    else:
      ":")
  let message = (label@info.style()) & " " & (info.message@messageStyle)
  let visualisation = info.visualize()
  let file =
    if info.kind == HighlightInfoKind.Error:
      stderr
    else:
      stdout

  file.write(message & "\n" & visualisation)

{.pop.} # raises: []

when isMainModule:
  proc getLine(buf: openArray[char]; n: Natural): string =
    proc skipUntilEndl(buf: openArray[char]; idx: var int) =
      while buf[idx] notin {'\r', '\n'}:
        idx += 1

    proc skipEndl(buf: openArray[char]; idx: var int) =
      if buf[idx] == '\r':
        idx += 1

      if buf[idx] == '\n':
        idx += 1

    var i = 0
    var lineNum = 1
    while lineNum < n:
      buf.skipUntilEndl(i)
      buf.skipEndl(i)
      lineNum += 1

    let start = i
    buf.skipUntilEndl(i)

    result = buf.toOpenArray(start, i - 1).substr()

  let filename = "tests_local/parser/struct.jet"
  let buf = readFile(filename)
  let range = FilePos(line: 3, column: 1) .. FilePos(line: 3, column: 5)

  highlightInfoInFile(HighlightInfo(
    message: "test message",
    target: some(HighlightTarget(
      range: range,
      line: buf.getLine(range.a.line.int),
      details: "test details",
    )),
    notes: @[
      HighlightNote(
        message: "this is note message",
        filePath: some(filename),
        range: some(range),
      ),
      HighlightNote(
        message: "another note",
      )
    ],
    filePath: filename,
    kind: HighlightInfoKind.Error,
    error: CompilationError.EMPTY,
  ))
