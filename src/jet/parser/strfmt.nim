import
  std/strformat,
  std/strutils,

  jet/ast,
  jet/literal,
  jet/lexer,
  jet/lexerbase,
  jet/parser,

  lib/lineinfo

{.push, raises: [].}

type
  FmtLexer* = object of LexerBase
  FmtLexerError* = object of CatchableError

const
  fmtSpecifierChars* = Letters + Digits + {'.', '_', '-', '+', '<', '>', '=', '!', '?'}

template raiseFmtLexerError(message: string) =
  raise (ref FmtLexerError)(msg: message)

func newFmtLexer*(node: AstNode): FmtLexer =
  assert(node.kind == Lit)
  assert(node.lit.kind == lkString)

  result = FmtLexer(
    buffer: node.lit.stringVal.toOpenArray(0, node.lit.stringVal.high),
    posOffset: node.rng.a.withOffset(1) - initialFilePos,
  )

func genFmtCall(result: var seq[AstNode]; expr, spec: string; exprPosOffset: FilePosition; specRange: FileRange)
  {.raises: [LexerError, ParserError, ValueError].} =
  # We procude code like this:
  # | $formatValue(`expr`, `spec`)
  let exprNode = parseExpr(expr, exprPosOffset)
  let specLit = initAstNodeLit(newLit(spec), specRange)
  let fmtFunc = initAstNodeBranch(Prefix, @[
    initAstNodeOperator(OpDollar),
    initAstNodeId("formatValue"),
  ])
  let fmtFuncArgs = initAstNodeBranch(List, @[exprNode, specLit])
  let formatValueCall = initAstNodeBranch(ExprRound, @[fmtFunc, fmtFuncArgs])
  result &= formatValueCall

func genBuf(result: var seq[AstNode]; buf: string; rng: FileRange) =
  let expr = initAstNodeLit(newLit(buf))
  result &= expr

func parseFmtString*(self: var FmtLexer): AstNode
  {.raises: [LexerError, FmtLexerError, ParserError, ValueError].} =
  var exprs = newSeq[AstNode]()
  var buf = ""
  var bufStartPos = emptyFilePos

  # TODO: lineinfo is broken in multiline literals

  while true:
    if self.popChar('$'):
      var expr = ""
      var spec = ""
      var exprPosOffset = emptyFilePos
      var specRange = emptyFileRange

      case self.peek()
      of '{':
        self.pop()
        exprPosOffset = self.peekPos() - initialFilePos
        expr = self.parseWhile(it notin {'}', ':'})

        if self.popChar(':'):
          let specStartPos = self.peekPos()
          spec = self.parseUntil(it == '}')
          specRange = specStartPos .. self.peekPos()

          if spec.len() == 0:
            raiseFmtLexerError("empty format specifier are not alloved")

          for c in spec:
            if c notin fmtSpecifierChars:
              raiseFmtLexerError(&"invalid character in format specifier: '{c}'")

        if not self.popChar('}'):
          raiseFmtLexerError("missing closing }")
      of IdStartChars:
        exprPosOffset = self.peekPos() - initialFilePos
        expr = self.parseWhile(it in IdChars)
      of '$':
        discard
      of '\0':
        raiseFmtLexerError("expected format specifier after '$', got end of string literal")
      else:
        raiseFmtLexerError(&"unexpected character after '$': '{self.peek()}'; for single '$' symbol write it twice")

      if expr.len() > 0:
        if buf.len() > 0:
          genBuf(exprs, move(buf), bufStartPos .. self.peekPos())
          buf = ""
          bufStartPos = emptyFilePos
        let posOffset = exprPosOffset
        genFmtCall(exprs, expr, spec, posOffset, specRange)

    if self.isEmpty():
      break

    if self.peek() in Newlines:
      self.handleNewLine()

    buf &= self.pop()

    if bufStartPos == emptyFilePos:
      bufStartPos = self.peekPos()

  if buf.len() > 0:
    genBuf(exprs, move(buf), bufStartPos .. self.peekPos())

  # TODO: prealloc string
  result = initAstNodeLit(newLit(""))

  for i, expr in exprs:
    result =
      if i == 0: expr
      else: initAstNodeBranch(Infix, @[
        initAstNodeOperator(OpAdd),
        result,
        expr,
      ])

func parseFmtString*(node: AstNode): AstNode
  {.raises: [LexerError, FmtLexerError, ParserError, ValueError].} =
  var lexer = newFmtLexer(node)
  result = lexer.parseFmtString()

{.pop.} # raises: []
