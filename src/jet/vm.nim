import std/sugar
import std/tables
import std/sequtils
import std/strutils
import std/strformat

import jet/ast
import jet/ast/sym
import jet/ast/types
import jet/literal
import jet/vm/obj
import jet/sem/typechecker

import lib/stack
import lib/utils
import lib/utils/text_style
import lib/utils/line_info


# TODO:
#   - some 'vmError' calls was replaces by 'panic'. This is invalid.
#   - implement missing features


# built-in
proc builtin_len(params: seq[Object]): Object
proc builtin_str(params: seq[Object]): Object
proc builtin_print(params: seq[Object]): Object
proc builtin_println(params: seq[Object]): Object
proc builtin_panic(params: seq[Object]): Object

let builtinSymTypes = [
    Type(
        name           : "",
        kind           : tyFunc,
        funcIsVarargs  : false,
        funcParams     : @[anyType],
        funcReturnType : i64Type),
    Type(
        name           : "",
        kind           : tyFunc,
        funcIsVarargs  : false,
        funcParams     : @[anyType],
        funcReturnType : stringType),
    Type(
        name           : "",
        kind           : tyFunc,
        funcIsVarargs  : true,
        funcParams     : @[anyType],
        funcReturnType : unitType),
    Type(
        name           : "",
        kind           : tyFunc,
        funcIsVarargs  : true,
        funcParams     : @[anyType],
        funcReturnType : unitType),
    Type(
        name           : "",
        kind           : tyFunc,
        funcIsVarargs  : true,
        funcParams     : @[anyType],
        funcReturnType : neverType),
]
let builtins = [
    Object(
        id       : "len",
        typ      : builtinSymTypes[0],
        kind     : skBuiltinFunc,
        oType    : tyFunc,
        fnScope  : nil,
        fnBody   : nil,
        fnParams : @[]),
    Object(
        id       : "str",
        typ      : builtinSymTypes[1],
        kind     : skBuiltinFunc,
        oType    : tyFunc,
        fnScope  : nil,
        fnBody   : nil,
        fnParams : @[]),
    Object(
        id       : "print",
        typ      : builtinSymTypes[2],
        kind     : skBuiltinFunc,
        oType    : tyFunc,
        fnScope  : nil,
        fnBody   : nil,
        fnParams : @[]),
    Object(
        id       : "println",
        typ      : builtinSymTypes[3],
        kind     : skBuiltinFunc,
        oType    : tyFunc,
        fnScope  : nil,
        fnBody   : nil,
        fnParams : @[]),
    Object(
        id       : "panic",
        typ      : builtinSymTypes[4],
        kind     : skBuiltinFunc,
        oType    : tyFunc,
        fnScope  : nil,
        fnBody   : nil,
        fnParams : @[]),
]
let unitObj = Object(
    id    : "",
    typ   : unitType,
    kind  : skLit,
    oType : tyUnit)
let nullObj = Object(
    id    : "",
    typ   : nullType,
    kind  : skLit,
    oType : tyNull)

type StackEntry = object
    tree*  {.requiresInit.} : Node
    scope* {.requiresInit.} : Scope
    info*  : LineInfo
    name*  : string

type VmState* = ref object
    rootNode*  {.requiresInit.} : Node
    rootScope* {.requiresInit.} : Scope

    info*  : LineInfo           ## Current line info
    tree*  : Node               ## Current node tree
    scope* : Scope              ## Current scope
    stack* : Stack[StackEntry]  ## VM stack


# ----- STACK ENTRY ----- #
proc newStackEntry(vm: VmState; tree: Node; scope: Scope; name: string; info: LineInfo): StackEntry =
    result = StackEntry(
        tree  : tree,
        scope : scope,
        name  : name,
        info  : info)

proc err(vm: VmState; msg: string) {.noreturn.}


# ----- SYMBOLS ----- #
proc lookupBuiltinFunc(sym: Sym): ((seq[Object]) {.nimcall.} -> Object) =
    result = nil

    if sym != nil and sym.kind == skBuiltinFunc:
        case sym.id
        of "len"     : result = builtin_len
        of "str"     : result = builtin_str
        of "print"   : result = builtin_print
        of "println" : result = builtin_println
        of "panic"   : result = builtin_panic
        else: discard


# ----- VM OPS ----- #
proc evalEqOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '==' op")

    let eq = case left.oType:
        of tyI8: left.i8Val == right.i8Val
        of tyI16: left.i16Val == right.i16Val
        of tyI32: left.i32Val == right.i32Val
        of tyI64: left.i64Val == right.i64Val
        of tyU8: left.u8Val == right.u8Val
        of tyU16: left.u16Val == right.u16Val
        of tyU32: left.u32Val == right.u32Val
        of tyU64: left.u64Val == right.u64Val
        of tyBool: left.boolVal == right.boolVal
        of tyChar: left.charVal == right.charVal
        else: unimplemented(fmt"'evalEqOp' for {left.oType}")
    result = Object(
        id      : "",
        typ     : boolType,
        kind    : skVal,
        oType   : tyBool,
        boolVal : eq)

proc evalNeOp(vm: VmState; left, right: Object): Object =
    result = vm.evalEqOp(left, right)
    result.boolVal = not result.boolVal

proc evalLtOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '<' op")

    let lt = case left.oType:
        of tyI8: left.i8Val < right.i8Val
        of tyI16: left.i16Val < right.i16Val
        of tyI32: left.i32Val < right.i32Val
        of tyI64: left.i64Val < right.i64Val
        of tyU8: left.u8Val < right.u8Val
        of tyU16: left.u16Val < right.u16Val
        of tyU32: left.u32Val < right.u32Val
        of tyU64: left.u64Val < right.u64Val
        of tyBool: left.boolVal < right.boolVal
        of tyChar: left.charVal < right.charVal
        else: unimplemented(fmt"'evalLtOp' for {left.oType}")
    result = Object(
        id      : "",
        typ     : boolType,
        kind    : skVal,
        oType   : tyBool,
        boolVal : lt)

proc evalGtOp(vm: VmState; left, right: Object): Object =
    result = vm.evalLtOp(right, left)

proc evalLeOp(vm: VmState; left, right: Object): Object =
    result = vm.evalGtOp(left, right)
    result.boolVal = not result.boolVal

proc evalGeOp(vm: VmState; left, right: Object): Object =
    result = vm.evalLtOp(left, right)
    result.boolVal = not result.boolVal

proc evalSumOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '+' op")

    result = case left.oType:
        of tyI8: Object(id: "", typ: i8Type, kind: skVal, oType: tyI8,
            i8Val: left.i8Val + right.i8Val)
        of tyI16: Object(id: "", typ: i16Type, kind: skVal, oType: tyI16,
            i16Val: left.i16Val + right.i16Val)
        of tyI32: Object(id: "", typ: i32Type, kind: skVal, oType: tyI32,
            i32Val: left.i32Val + right.i32Val)
        of tyI64: Object(id: "", typ: i64Type, kind: skVal, oType: tyI64,
            i64Val: left.i64Val + right.i64Val)
        of tyU8: Object(id: "", typ: u8Type, kind: skVal, oType: tyU8,
            u8Val: left.u8Val + right.u8Val)
        of tyU16: Object(id: "", typ: u16Type, kind: skVal, oType: tyU16,
            u16Val: left.u16Val + right.u16Val)
        of tyU32: Object(id: "", typ: u32Type, kind: skVal, oType: tyU32,
            u32Val: left.u32Val + right.u32Val)
        of tyU64: Object(id: "", typ: u64Type, kind: skVal, oType: tyU64,
            u64Val: left.u64Val + right.u64Val)
        else: unimplemented(fmt"'evalSumOp' for {left.oType}")

proc evalSubOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '-' op")

    result = case left.oType:
        of tyI8: Object(id: "", typ: i8Type, kind: skVal, oType: tyI8,
            i8Val: left.i8Val - right.i8Val)
        of tyI16: Object(id: "", typ: i16Type, kind: skVal, oType: tyI16,
            i16Val: left.i16Val - right.i16Val)
        of tyI32: Object(id: "", typ: i32Type, kind: skVal, oType: tyI32,
            i32Val: left.i32Val - right.i32Val)
        of tyI64: Object(id: "", typ: i64Type, kind: skVal, oType: tyI64,
            i64Val: left.i64Val - right.i64Val)
        of tyU8: Object(id: "", typ: u8Type, kind: skVal, oType: tyU8,
            u8Val: left.u8Val - right.u8Val)
        of tyU16: Object(id: "", typ: u16Type, kind: skVal, oType: tyU16,
            u16Val: left.u16Val - right.u16Val)
        of tyU32: Object(id: "", typ: u32Type, kind: skVal, oType: tyU32,
            u32Val: left.u32Val - right.u32Val)
        of tyU64: Object(id: "", typ: u64Type, kind: skVal, oType: tyU64,
            u64Val: left.u64Val - right.u64Val)
        else: unimplemented(fmt"'evalSumOp' for {left.oType}")

proc evalMulOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '*' op")

    result = case left.oType:
        of tyI8: Object(id: "", typ: i8Type, kind: skVal, oType: tyI8,
            i8Val: left.i8Val * right.i8Val)
        of tyI16: Object(id: "", typ: i16Type, kind: skVal, oType: tyI16,
            i16Val: left.i16Val * right.i16Val)
        of tyI32: Object(id: "", typ: i32Type, kind: skVal, oType: tyI32,
            i32Val: left.i32Val * right.i32Val)
        of tyI64: Object(id: "", typ: i64Type, kind: skVal, oType: tyI64,
            i64Val: left.i64Val * right.i64Val)
        of tyU8: Object(id: "", typ: u8Type, kind: skVal, oType: tyU8,
            u8Val: left.u8Val * right.u8Val)
        of tyU16: Object(id: "", typ: u16Type, kind: skVal, oType: tyU16,
            u16Val: left.u16Val * right.u16Val)
        of tyU32: Object(id: "", typ: u32Type, kind: skVal, oType: tyU32,
            u32Val: left.u32Val * right.u32Val)
        of tyU64: Object(id: "", typ: u64Type, kind: skVal, oType: tyU64,
            u64Val: left.u64Val * right.u64Val)
        else: unimplemented(fmt"'evalSumOp' for {left.oType}")

proc evalDivOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '/' op")

    result = case left.oType:
        of tyI8: Object(id: "", typ: i8Type, kind: skVal, oType: tyI8,
            i8Val: left.i8Val div right.i8Val)
        of tyI16: Object(id: "", typ: i16Type, kind: skVal, oType: tyI16,
            i16Val: left.i16Val div right.i16Val)
        of tyI32: Object(id: "", typ: i32Type, kind: skVal, oType: tyI32,
            i32Val: left.i32Val div right.i32Val)
        of tyI64: Object(id: "", typ: i64Type, kind: skVal, oType: tyI64,
            i64Val: left.i64Val div right.i64Val)
        of tyU8: Object(id: "", typ: u8Type, kind: skVal, oType: tyU8,
            u8Val: left.u8Val div right.u8Val)
        of tyU16: Object(id: "", typ: u16Type, kind: skVal, oType: tyU16,
            u16Val: left.u16Val div right.u16Val)
        of tyU32: Object(id: "", typ: u32Type, kind: skVal, oType: tyU32,
            u32Val: left.u32Val div right.u32Val)
        of tyU64: Object(id: "", typ: u64Type, kind: skVal, oType: tyU64,
            u64Val: left.u64Val div right.u64Val)
        else: unimplemented(fmt"'evalSumOp' for {left.oType}")

proc evalConcatOp(vm: VmState; left, right: Object): Object =
    if not isEqTypes(left.typ, right.typ):
        vm.err(fmt"type mismatch, got {left.oType} and {right.oType} for '++' op")

    result = case left.oType:
        of tyString: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: left.stringVal & right.stringVal)
        else:
            panic("")
            # errTypeMismatch($left.oType, fmt"one of [{tyString}]")


# ----- NEW VM ----- #
proc injectArgsIntoScope(fn: Object; args: seq[Object]): Scope =
    ## `args.len` must be the same with `fn.fnParams.len`.
    result = newScopeEnclosed(fn.fnScope)
    var i  = 0

    for param in fn.fnParams:
        result[param.id] = args[i]
        inc(i)

proc wrapReturnVal(obj: Object): Object
proc wrapReturnValMaybe(obj: Object): Object
proc unwrapReturnVal(obj: var Object)
proc unwrapReturnValMaybe(obj: var Object)

proc newVm*(tree: Node): VmState
proc eval*(vm: VmState; tree: Node): Object
proc eval*(vm: VmState): Object
proc evalProgram(vm: VmState): Object
proc evalCall(vm: VmState): Object
proc evalLit(vm: VmState): Object
proc evalLetStmt(vm: VmState): Object
proc evalDefStmt(vm: VmState): Object
proc evalReturnStmt(vm: VmState): Object
proc evalBlock(vm: VmState): Object
proc evalBlockWithScope(vm: VmState; scope: Scope): Object
proc evalId(vm: VmState): Object
proc evalInfix(vm: VmState): Object
proc evalIfExpr(vm: VmState): Object


proc newVm(tree: Node): VmState =
    result = VmState(rootNode: tree, rootScope: newScope(), info: LineInfo())
    result.tree  = result.rootNode
    result.scope = result.rootScope

proc eval(vm: VmState; tree: Node): Object =
    vm.tree = tree

    result = case tree.kind:
        of nkEmpty      : nullObj
        of nkProgram    : vm.evalProgram()
        of nkExprParen  : vm.evalCall()
        of nkLit        : vm.evalLit()
        of nkVarDecl    : vm.evalLetStmt()
        of nkDefStmt    : vm.evalDefStmt()
        of nkReturnStmt : vm.evalReturnStmt()
        of nkIfExpr     : vm.evalIfExpr()
        of nkElseBranch : vm.evalBlock()
        of nkEqExpr     : vm.evalBlock()
        of nkDoExpr     : vm.evalBlock()
        of nkId         : vm.evalId()
        of nkInfix      : vm.evalInfix()
        else: vm.err(fmt"unimplemented: eval '{$tree.kind}'")

proc eval(vm: VmState): Object =
    assert(vm.rootNode.kind == nkProgram)
    vm.eval(vm.rootNode)

proc evalProgram(vm: VmState): Object =
    result = nullObj
    let programTree = vm.tree
    for stmt in programTree.children:
        result = vm.eval(stmt)

proc evalCall(vm: VmState): Object =
    let entry = vm.newStackEntry(vm.tree, vm.rootScope, vm.tree[0].id, LineInfo())
    vm.stack.push(entry)

    let name = vm.tree[0].id
    let args = vm.tree[1].children.mapIt(vm.eval(it))

    # debug fmt"call '{name}', args: {args}"

    var fnObj = nil.Object

    block BuiltInLookup:
        let builtIn = builtIns.filterIt(it.id == name)

        if builtIn.len() == 1:
            fnObj = builtIn[0]

    if fnObj == nil:
        fnObj = vm.scope[name].Object
    if fnObj == nil:
        vm.err(fmt"identifier '{name}' is undefined")
    if fnObj.oType != tyFunc:
        vm.err(fmt"identifier '{name}' is not a function")

    let invalid = checkFuncTypes(fnObj.typ, args.mapIt(it.typ))
    if invalid != -1:
        if invalid < fnObj.fnParams.len():
            let paramName = $fnObj.fnParams[invalid].id

            vm.err(
                fmt"invalid type for {invalid + 1} parameter " &
                fmt"'{paramName}', " &
                fmt"got {$args[invalid].typ.kind}, " &
                fmt"expected {$fnObj.typ.funcParams[invalid].kind}")
        else:
            vm.err(fmt"missing {invalid + 1} argument for '{fnObj.id}', expected {$fnObj.typ.funcParams[invalid].kind}")

    if fnObj.kind == skBuiltinFunc:
        let fn = lookupBuiltinFunc(fnObj)
        result = fn(args)
    elif fnObj.kind == skFunc:
        let fnScope = fnObj.injectArgsIntoScope(args)
        let oldTree = vm.tree
        vm.tree = fnObj.fnBody
        result = vm.evalBlockWithScope(fnScope)
        vm.tree = oldTree
        unwrapReturnVal(result)
    else:
        unreachable()

proc evalLit(vm: VmState): Object =
    let lit = vm.tree.lit

    result = case lit.kind:
        of tlkString: Object(
            typ       : stringType,
            kind      : skLit,
            oType     : tyString,
            stringVal : lit.stringVal)
        of tlkI32: Object(
            typ    : i32Type,
            kind   : skLit,
            oType  : tyI32,
            i32Val : lit.i32Val)
        of tlkI64: Object(
            typ    : i64Type,
            kind   : skLit,
            oType  : tyI64,
            i64Val : lit.i64Val)
        of tlkBool: Object(
            typ     : boolType,
            kind    : skLit,
            oType   : tyBool,
            boolVal : lit.boolVal)
        else: unimplemented(fmt"'evalLit' for {lit.kind}")

proc evalLetStmt(vm: VmState): Object =
    result = unitObj

    let names = vm.tree[0 ..^ 4].mapIt(it.id)
    let typ   = vm.tree[^3]
    let expr  = vm.eval(vm.tree[^2])

    for name in names:
        if typ.kind != nkEmpty:
            let expectedType = block:
                let builtInTypes = builtInTypes.filterIt(it.name == typ.id)

                if builtInTypes.len() < 1: vm.scope[typ.id].typ
                elif builtInTypes.len() > 1: unreachable()
                else: builtInTypes[0]
            if expr.typ != expectedType:
                vm.err(fmt"type mismatch, got {expr.typ}, expected {expectedType}")

        if vm.scope[name] != nil:
            vm.err(fmt"identifier '{name}' is already defined")

        let variable = Object(
            id    : name,
            typ   : expr.typ,
            kind  : skVar,
            oType : expr.typ.kind
        )
        variable.copyVal(expr)
        vm.scope[name] = variable

proc evalDefStmt(vm: VmState): Object =
    if vm.tree[0].kind == nkExprDotExpr:
        unimplemented("vm methods")

    result = unitObj

    let name    = vm.tree[0].id
    let body    = vm.tree[3]
    let fnScope = newScopeEnclosed(vm.scope.clone())
    var params  = newSeq[Sym]()

    for param in vm.tree[1].children:
        for paramId in param[0 ..^ 4].mapIt(it.id):
            if param[^3].kind == nkEmpty:
                vm.err(fmt"VM cannot infer type for parameters")

            let paramTypeId = param[^3].id
            var paramType   = types.builtinTypes.filterIt(it.name == paramTypeId)[0]

            if paramType == nil:
                let paramTypeSym = vm.scope[paramTypeId]

                if paramTypeSym == nil:
                    vm.err(fmt"identifier '{paramTypeId}' is undefined")
                if paramTypeSym.kind != skType:
                    vm.err(fmt"identifier '{paramTypeId}' is not a type")

                paramType = paramTypeSym.typ

            params &= Sym(
                id   : paramId,
                typ  : paramType,
                kind : skParam)

    let returnTypeId = vm.tree[2].id
    var returnType   = types.builtinTypes.filterIt(it.name == returnTypeId)[0]

    if returnType == nil:
        let returnTypeSym = vm.scope[returnTypeId]

        if returnTypeSym == nil:
            vm.err(fmt"identifier '{returnTypeId}' is undefined")
        if returnTypeSym.kind != skType:
            vm.err(fmt"identifier '{returnTypeId}' is not a type")

        returnType = returnTypeSym.typ

    let fnType = Type(
        name           : name,
        kind           : tyFunc,
        funcIsVarargs  : false,
        funcParams     : params.mapIt(it.typ),
        funcReturnType : returnType)
    let fnObj = Object(
        id       : name,
        typ      : fnType,
        kind     : skFunc,
        oType    : tyFunc,
        fnParams : params,
        fnBody   : body,
        fnScope  : fnScope)

    fnScope[name]  = fnObj # for recursion
    vm.scope[name] = fnObj

proc evalReturnStmt(vm: VmState): Object =
    result = vm.eval(vm.tree[0]).wrapReturnValMaybe()

proc evalBlock(vm: VmState): Object =
    result = unitObj
    let blockTree = vm.tree

    for stmt in blockTree.children:
        result = vm.eval(stmt)

        if result.kind == skReturnVal:
            break

proc evalBlockWithScope(vm: VmState; scope: Scope): Object =
    let oldScope = vm.scope
    vm.scope = scope
    result = vm.evalBlock()
    vm.scope = oldScope

proc evalId(vm: VmState): Object =
    let name = vm.tree.id
    let sym  = block:
        let builtInSyms = builtins.filterIt(it.id == name)

        if builtInSyms.len() < 1: vm.scope[name].Object
        elif builtInSyms.len() > 1: unreachable()
        else: builtInSyms[0]

    if sym == nil:
        vm.err(fmt"identifier '{name}' is undefined")

    result = Object(id: name, kind: sym.kind, typ: sym.typ, oType: sym.oType)
    result.copyVal(sym)

proc evalInfix(vm: VmState): Object =
    let tree  = vm.tree
    let op    = vm.tree[0].id
    var left  = vm.eval(tree[1])
    var right = vm.eval(tree[2])

    vm.tree = tree
    left.unwrapReturnValMaybe()
    right.unwrapReturnValMaybe()

    assert(left != nil)
    assert(right != nil)

    result = case op:
        of "==" : vm.evalEqOp(left, right)
        of "!=" : vm.evalNeOp(left, right)
        of "<"  : vm.evalLtOp(left, right)
        of "<=" : vm.evalLeOp(left, right)
        of ">"  : vm.evalGtOp(left, right)
        of ">=" : vm.evalGeOp(left, right)
        of "+"  : vm.evalSumOp(left, right)
        of "-"  : vm.evalSubOp(left, right)
        of "*"  : vm.evalMulOp(left, right)
        of "/"  : vm.evalDivOp(left, right)
        of "++" : vm.evalConcatOp(left, right)
        else: vm.err(fmt"invalid operator '{op}'")

proc evalIfExprAux(vm: VmState; branches: openArray[Node]): int =
    result = -1

    for i, branch in branches:
        case branch.kind
        of nkIfBranch:
            let cond = vm.eval(branch[0])

            if not isEqTypes(cond.typ, boolType):
                vm.err(fmt"type mismatch, expected bool for if condition, got {$cond.typ}")
            if cond.boolVal:
                return i
        of nkElseBranch:
            assert(i == branches.high, "illformed AST; else branch must be the last branch in the 'IfExpr' tree")
            return i
        else: unreachable("Illformed AST")

proc evalIfExpr(vm: VmState): Object =
    assert(vm.tree.len() > 0, "Illformed AST")
    result = unitObj

    # TODO: check branch types (semantic check stage)
    let oldTree   = vm.tree
    let branchIdx = vm.evalIfExprAux(vm.tree.children)
    vm.tree = oldTree

    if branchIdx != -1:
        let body = block:
            let tree = vm.tree[branchIdx]
            if tree.kind != nkIfBranch:
                assert(tree.kind == nkElseBranch)
                tree
            else: tree[1]
        result = vm.eval(body)
    else:
        result = unitObj

proc entryInfo(entry: StackEntry): string =
    result = fmt"in '{entry.name}' at {entry.info}"

proc unwrapStack(vm: VmState): seq[string] =
    result = @[]
    for entry in vm.stack.poppedItems():
        result.add(entry.entryInfo())

proc err(vm: VmState; msg: string) =
    let stack = vm.unwrapStack()
    panic(msg & "\n  " & stack.join("\n  "), vm.info)


# ----- BUILT-INS ----- #
proc builtin_len(params: seq[Object]): Object =
    if params.len() != 1:
        unreachable(fmt"typechecker is dumb, expected 1 argument, got {params.len()}")

    var param = params[0]
    param.unwrapReturnValMaybe()

    result = case param.oType:
        of tyString: Object(
            id: "", typ: usizeType, kind: skVal,
            oType: tyUSize, usizeVal: param.stringVal.len().uint)
        else: panic(fmt"type mismatch, got {param.oType} for builtin_len, expected string")

proc builtin_str(params: seq[Object]): Object =
    if params.len() != 1:
        panic(fmt"invalid count of arguments, got {params.len()}, expected 1")

    var param = params[0]
    param.unwrapReturnValMaybe()

    result = case param.oType:
        of tyBool: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.boolVal)
        of tyString: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: param.stringVal)
        of tyISize: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.isizeVal)
        of tyUSize: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.usizeVal)
        of tyI8: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.i8Val)
        of tyI16: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.i16Val)
        of tyI32: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.i32Val)
        of tyI64: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.i64Val)
        of tyU8: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.u8Val)
        of tyU16: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.u16Val)
        of tyU32: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.u32Val)
        of tyU64: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.u64Val)
        of tyF32: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.f32Val)
        of tyF64: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: $param.f64Val)
        of tyFunc: Object(
            id: "", typ: stringType, kind: skVal,
            oType: tyString, stringVal: "func(...)")
        of tyStruct:
            var fields = newSeq[string]()
            for id, val in param.structFields.pairs():
                fields &= fmt"{id} = {val}"
            let fieldsStr = fields.join(", ")
            let str = fmt"{param.typ.name} {{ {fieldsStr} }}"
            Object(
                id: "", typ: stringType, kind: skVal,
                oType: tyString, stringVal: str)
        of tyEnum:
            let str = fmt"{param.typ.name}.{param.typ.enumFields[param.enumVal]}"
            Object(
                id: "", typ: stringType, kind: skVal,
                oType: tyString, stringVal: str)
        else: unimplemented(fmt"'builtin_to_str' for {param.oType}")

proc builtin_print(params: seq[Object]): Object =
    const style = TextStyle(foreground: BrightBlue)

    var msg = ""
    for param in params:
        msg &= builtin_str(@[param]).stringVal

    stdout.write(stylizeText("> " & msg, style))
    result = unitObj

proc builtin_println(params: seq[Object]): Object =
    result = builtin_print(params)
    stdout.write('\n')

proc builtin_panic(params: seq[Object]): Object =
    const panicTagStyle = TextStyle(foreground: Red, bold: true)
    const style         = TextStyle(foreground: BrightRed)

    var msg = ""
    for param in params:
        msg &= builtin_str(@[param]).stringVal

    result = nil # never

    stdout.write("panic! " @ panicTagStyle)
    stdout.writeLine(msg @ style)
    quit(QuitFailure)


# ----- UTILS ----- #
proc wrapReturnVal(obj: Object): Object =
    ## **Returns:** a new object with the same value with `obj`,
    ## but with kind `skVal`.
    result = Object(
        id    : obj.id,
        typ   : obj.typ,
        kind  : skReturnVal,
        oType : obj.oType)
    result.copyVal(obj)

proc wrapReturnValMaybe(obj: Object): Object =
    if obj.kind != skReturnVal:
        result = wrapReturnVal(obj)
    else:
        result = obj

proc unwrapReturnVal(obj: var Object) =
    assert(obj != nil)

    if obj.kind != skReturnVal and obj.oType notin {tyUnit, tyNever}:
        panic(fmt"expected return value, got {obj.kind}")

    let tmp = Object(
        id    : obj.id,
        typ   : obj.typ,
        kind  : obj.kind,
        oType : obj.oType)
    tmp.copyVal(obj)
    obj = Object(
        id    : obj.id,
        typ   : obj.typ,
        kind  : skVal,
        oType : obj.oType)
    obj.copyVal(tmp)

proc unwrapReturnValMaybe(obj: var Object) =
    assert(obj != nil)

    if obj.kind == skReturnVal:
        unwrapReturnVal(obj)
