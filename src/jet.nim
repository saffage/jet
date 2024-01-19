import
  std/strformat,
  std/strutils,
  std/sequtils,
  std/os,
  std/options,

  jet/astalgo,
  jet/token,
  jet/lexer,
  jet/parser,
  jet/symbol,
  jet/module,
  jet/sem,

  lib/utils,

  pkg/results

proc main() =
  logger.maxErrors = 3

  if not dirExists(getAppDir().parentDir() / "lib"):
    panic("can't find core library directory: \"$jet/lib\"")

  # Pipeline:
  #   - tokenize
  #   - parse AST
  #   - (?) annonations resolve
  #   - semantic checks
  #   - (?) typed AST
  #   - (?) deffered annonations resolve (typed annonations)
  #   - backend stage

  if paramCount() != 1:
    panic("expected path to Jet file as 1 argument")

  hint("file reading...")
  let argument = paramStr(1)
  let file     = open(argument, fmRead).readAll()

  hint("lexical analysis...")
  var tokens = try:
    var lexer = newLexer(file)
    lexer.getAllTokens()
  except LexerError as e:
    error(e.msg, e.info)
    raise
  let tmp1 = "  " & tokens.mapIt(it.human()).join("\n  ")
  debug(&"tokens: \n{tmp1}")

  hint("normalizing tokens...")
  tokens = normalizeTokens(tokens)
  let tmp2 = "  " & tokens.mapIt(it.human()).join("\n  ")
  debug(&"normalized tokens: \n{tmp2}")

  hint("syntactic analysis...")
  var parser = newParser(tokens)
  try:
    parser.parseAll()
  except ParserError as e:
    error(e.msg, e.info)
    raise

  hint("done")

  if parser.getAst().isSome():
    debug("generated AST")
    parser.getAst().get().printTree()
  else:
    debug("AST is not generated")

  hint("semantic analysis...")
  var rootTree = parser.getAst().get()
  var mainModule = newModule(rootTree)
  mainModule.traverseSymbols()
  echo mainModule.rootScope.repr

when isMainModule: main()
