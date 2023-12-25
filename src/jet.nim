import std/os
import std/tables
import std/sequtils
import std/strutils
import std/strformat

import jet/scanner
import jet/parser
import jet/ast
import jet/ast/sym
import jet/vm
import jet/vm/obj

import utils
import utils/text_style


# proc print_node_tree(node: Node) {.importc: "print_node_tree", dynlib: "vm_test.dll", cdecl.}

proc main() =
    const typeStyle = TextStyle(foreground: BrightCyan, italic: true)

    logger.maxErrors = 3
    when defined(release): logger.loggingLevel = Error

    if not dirExists(getAppDir().parentDir() / "lib"):
        panic("can't find core library directory: \"$jet/lib\"")

    if not dirExists(getAppDir().parentDir() / "lib" / "std"):
        panic("can't find STD library directory: \"$jet/lib/std\"")

    # Pipeline:
    #   - tokenize
    #   - parse to AST
    #   - (?) pragma resolve
    #   - semantic checks
    #   - (?) typed AST
    #   - (?) deffered pragma resolve (typed pragmas)
    #   - backend stage
    #
    # Backends:
    #   - C (WIP)
    #   - Jet VM
    #   - Bizzare VM (JIT, WIP)

    if paramCount() > 0:
        let argument  = paramStr(1)
        var scanner   = openScannerFile(argument)
        var parser    = newParser(scanner)
        var program   = parser.parseAll()

        echo(program.treeRepr)

        var vm        = newVm(program)
        let evaluated = vm.eval()

        unreachable(evaluated != nil, "eval result can't be null")

        # if getCursorPos().x > 0: echo()
        echo stylizeText(evaluated.inspect(), typeStyle)

        let objs = $vm.scope.syms.keys().toSeq().join(", ")
        echo stylizeText(fmt"scope = {{ {objs} }}", typeStyle)

        # echo "call 'vm_test'"
        # print_node_tree(program)
    else:
        unimplemented("REPL")

when isMainModule: main()
