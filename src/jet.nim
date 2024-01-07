import std/os
# import std/tables
# import std/sequtils
# import std/strutils
# import std/strformat

import jet/lexer
import jet/parser
import jet/ast
# import jet/ast/algo
# import jet/ast/sym
# import jet/sem
# import jet/token
# import jet/vm
# import jet/vm/obj

import lib/utils
# import lib/utils/text_style


proc main() =
    # const typeStyle = TextStyle(foreground: BrightCyan, italic: true)

    logger.maxErrors = 3
    when defined(release): logger.loggingLevel = Error

    if not dirExists(getAppDir().parentDir() / "lib"):
        panic("can't find core library directory: \"$jet/lib\"")

    if not dirExists(getAppDir().parentDir() / "lib" / "std"):
        panic("can't find STD library directory: \"$jet/lib/std\"")

    # Pipeline:
    #   - tokenize
    #   - parse AST
    #   - (?) annonations resolve
    #   - semantic checks
    #   - (?) typed AST
    #   - (?) deffered annonations resolve (typed annonations)
    #   - backend stage
    #
    # Backends:
    #   - C
    #   - Jet VM

    if paramCount() > 0:
        let argument = paramStr(1)
        var lexer    = newLexerFromFileName(argument)
        var parser   = newParser(lexer)
        var program  = parser.parseAll()

        echo(program.treeRepr)
        # echo("Recreated AST:")
        # echo(ast2jet(program))

        # semGenTreeTypes(program)
        # semGenTreeSyms(program)

        # var vm        = newVm(program)
        # let evaluated = vm.eval()

        # assert(evaluated != nil, "eval result can't be null")

        # # if getCursorPos().x > 0: echo()
        # echo stylizeText(evaluated.inspect(), typeStyle)

        # let objs = $vm.scope.syms.keys().toSeq().join(", ")
        # echo stylizeText(fmt"scope = {{ {objs} }}", typeStyle)

        # echo "call 'vm_test'"
        # print_node_tree(program)
    else:
        unimplemented("REPL")

when isMainModule: main()
