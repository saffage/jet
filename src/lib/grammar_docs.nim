import std/macros
import std/strutils


var
    grammarBuffer {.compileTime.} = ""

proc getAnnotation(line: string): string
    {.compileTime.} =
    if line.len() < 2 or line[0] != '@' or line[1] notin IdentStartChars:
        result = ""
    else:
        result = $line[1]
        var i  = 2

        while i < line.len():
            if (let c = line[i]; c in IdentChars):
                result.add(c)
                inc(i)
            else:
                break

proc warningWithLineInfo(ident: string; lineinfo: sink LineInfo; lineOffset: int): NimNode =
    lineinfo.line += lineOffset
    lineinfo.column += 3

    result = ident(ident)
    result.setLineInfo(lineinfo)

proc getGrammar*(): string
    {.compileTime.} =
    ## **Returns:** all currently builded grammar
    return grammarBuffer

when not defined(jetGrammarDocs):
    macro grammarDocs*(node: typed) = node
else:
    macro grammarDocs*(node: typed) =
        ## Pragma to scan comment of the procedure and extract a
        ## grammar from it.
        ## To mark a block of grammar write `@grammar` in the beginnig
        ## of a grammar block and `@end` in the end.
        ##
        ## To build grammar just call `getGrammar()` proc.
        case node.kind
        of RoutineNodes:
            if node.body.kind == nnkEmpty:
                warning("Can't get a comment (declarations are not supported)", node)
                return node
            elif node.body[0].kind != nnkCommentStmt:
                warning("No comments provided for building the grammar", node)
            else:
                let comment        = node.body[0]
                var inGrammarBlock = false
                var wasWritten     = false
                var openLineOffset = 0

                for lineOffset, line in splitLines($comment).pairs():
                    var annotation = getAnnotation(line)

                    if annotation.cmpIgnoreStyle("grammar") == 0:
                        inGrammarBlock = true
                        openLineOffset = lineOffset
                        continue
                    elif annotation.cmpIgnoreStyle("end") == 0:
                        if not inGrammarBlock:
                            warning("Missing opening '@grammar'", warningWithLineInfo("end", comment.lineInfoObj, lineOffset))
                            continue
                        inGrammarBlock = false
                        continue

                    if inGrammarBlock:
                        wasWritten = true
                        grammarBuffer.add(line & '\n')

                if inGrammarBlock:
                    warning("Missing closing '@end'", warningWithLineInfo("grammar", comment.lineInfoObj, openLineOffset))

                if not wasWritten:
                    warning("Missing or empty grammar block in the comment", comment)
                else:
                    grammarBuffer.add('\n')
            result = node
        of nnkStmtList:
            if node.len() > 1 or node[0].kind != nnkCommentStmt:
                let errNode =
                    if node.len() <= 1: node
                    else: node[1]
                error("The only comment statement accepted for 'StmtList' for building grammar from it", errNode)
            grammarBuffer.add(node[0].strVal & "\n\n")
            result = newEmptyNode()
        else:
            error("Invalid node to build grammar", node)
            result = node


when isMainModule:
    proc parseX() {.used, grammarDocs.} =
        ## @grammar
        ## X <- ...
        ## @end
        discard

    grammarDocs do:
        ## Global grammar

    proc parseY() {.used, grammarDocs.} =
        ## @grammar
        ## Y <- ...
        ## @end
        discard

    const grammar = getGrammar()
    echo '\"' & grammar & '\"'
