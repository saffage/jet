import std/os

import jet/lexer
import jet/parser
import jet/ast

import lib/utils


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
        quit("expected path to Jet file as 1 argument")

    let argument = paramStr(1)
    var lexer    = newLexerFromFileName(argument)
    var parser   = newParser(lexer)
    var program  = parser.parseAll()

    # for token in lexer.getAllTokens():
        # echo(token.human())

    echo(program.treeRepr)

when isMainModule: main()
