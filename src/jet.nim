import
  std/strformat,
  std/strutils,
  std/sequtils,
  std/os,
  std/options,
  std/json,
  std/parseopt,

  jet/astalgo,
  jet/token,
  jet/lexer,
  jet/parser,
  jet/symbol,
  jet/module,
  jet/sem,

  lib/utils,
  lib/lineinfo

# WHY???
proc `%`*(v: char): JsonNode =
  result = JsonNode(kind: JString, str: $v)

proc main() =
  logger.maxErrors = 3

  let libDir = getAppDir().parentDir() / "lib"

  if not dirExists(libDir):
    panic("can't find core library directory: \"$jet/lib\"")

  # Pipeline:
  #   - read file
  #   - lexical analysis
  #   - normalizing tokens
  #   - syntactic analysis
  #   - generate AST
  #   - (?) resolve macros
  #   - semantic analysis
  #   - generate typed AST
  #   - (?) resolve annotations
  #   - backend stage

  let params   = commandLineParams()
  var argument = ""
  var noSem    = false

  for kind, key, val in params.getopt():
    case kind
    of cmdLongOption, cmdShortOption:
      case key
      of "nosem":
        noSem = val == "" or val.parseBool()
      else:
        panic("unknown option: " & key)
    of cmdArgument:
      argument = key
    of cmdEnd: discard

  if argument == "":
    panic("expected path to Jet file")

  hint("file reading...")
  let file = open(argument, fmRead).readAll()

  hint("lexical analysis...")
  var tokens = try:
    var lexer = newLexer(file)
    lexer.getAllTokens()
  except LexerError as e:
    # TODO: file id
    stdout.write(argument & ":" & $e.rng.a & ": ")
    error(e.msg)
    raise
  let tmp1 = "  " & tokens.mapIt(it.human()).join("\n  ")
  debug(&"tokens: \n{tmp1}")

  hint("normalizing tokens...")
  tokens = normalizeTokens(tokens)
  let tmp2 = "  " & tokens.mapIt(it.human()).join("\n  ")
  debug(&"normalized tokens: \n{tmp2}")

  hint("syntactic analysis...")
  var parser = newParser(tokens, filename=argument)
  try:
    parser.parseAll()
  except ParserError as e:
    # TODO: file id
    stdout.write(argument & ":" & $e.rng.a & ": ")
    error(e.msg)
    raise

  hint("done")

  if parser.getAst().isSome():
    debug("generated AST")
    parser.getAst().get().printTree()
    writeFile("tests_local/ast.json", (%parser.getAst().get()).pretty())
  else:
    debug("AST is not generated")

  if not noSem:
    hint("semantic analysis...")
    var rootTree = parser.getAst()
    var mainModule = newModule(rootTree)
    try:
      mainModule.traverseSymbols()
      debug("Root scope symbols:\n    " & mainModule.scope.symbols.join("\n    "))
    except SemanticError, ModuleError, ValueError:
      let err = getCurrentException()
      if err of (ref SemanticError):
        let err = (ref SemanticError)(err)
      raise

when isMainModule:
  main()
