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

import utils
import utils/text_style
import utils/line_info

import lib/stack


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

type VmError* = object of CatchableError


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
        of nkLetStmt    : vm.evalLetStmt()
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

    debug fmt"call '{name}', args: {args}"

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
    let name = vm.tree[0].id
    let typ  = vm.tree[1]
    let expr = vm.eval(vm.tree[2])

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

    result = unitObj

proc evalDefStmt(vm: VmState): Object =
    if vm.tree[0].kind == nkExprDotExpr:
        unimplemented("vm methods")

    result = unitObj

    let name    = vm.tree[0].id
    let body    = vm.tree[3]
    let fnScope = newScopeEnclosed(vm.scope.clone())
    var params  = newSeq[Sym]()

    for param in vm.tree[1].children:
        let paramId     = param[0].id
        let paramTypeId = param[1].id
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
            assert(i == branches.high, "else branch must be the last branch in the IfExpr AST")
            return i
        else: unreachable("Illformed AST")

proc evalIfExpr(vm: VmState): Object =
    assert(vm.tree.len() > 0, "Illformed AST")
    result = unitObj

    let oldTree   = vm.tree
    let branchIdx = vm.evalIfExprAux(vm.tree.children)
    vm.tree = oldTree

    let body = block:
        let tree = vm.tree[branchIdx]
        if tree.kind == nkIfBranch:
            tree[1]
        else:
            assert(tree.kind == nkElseBranch)
            tree

    result = vm.eval(body)

proc entryInfo(entry: StackEntry): string =
    result = fmt"in '{entry.name}' at {entry.info}"

proc unwrapStack(vm: VmState): seq[string] =
    result = @[]
    for entry in vm.stack.poppedItems():
        result.add(entry.entryInfo())

proc err(vm: VmState; msg: string) =
    let stack = vm.unwrapStack()
    panic(msg & "\n  " & stack.join("\n  "), vm.info)


# ----- VM ----- #
proc errTypeMismatch(got, expected: string) {.noreturn.}
proc errTypeMismatchFor(forX, got, expected: string) {.noreturn.}
proc vmError(msg: string) {.noreturn.}

proc eval*(node: Node; scope: Scope): Object
proc evalProgram(program: Node; scope: Scope): Object
proc evalCall(node: Node; scope: Scope): Object
proc evalLit(node: Node; scope: Scope): Object
proc evalFunc(node: Node; scope: Scope): Object
# proc evalEnum(node: Node; scope: Scope): Object
proc evalStruct(node: Node; scope: Scope): Object
proc evalStructInit(node: Node; scope: Scope): Object
proc evalBlock(node: Node; scope: Scope): Object
proc evalAssign(node: Node; scope: Scope): Object
proc evalId(node: Node; scope: Scope): Object
# proc evalIfExpr(node: Node; scope: Scope): Object
proc evalWhile(node: Node; scope: Scope): Object
proc evalReturn(node: Node; scope: Scope): Object
proc evalLet(node: Node; scope: Scope): Object
# proc evalInfixOp(node: Node; scope: Scope): Object
proc evalDotExpr(node: Node; scope: Scope): Object
proc evalDoExpr(node: Node; scope: Scope): Object


proc errTypeMismatch(got, expected: string) =
    vmError(fmt"invalid type, got {got}, expected {expected}")

proc errTypeMismatchFor(forX, got, expected: string) =
    vmError(fmt"invalid type for {forX}, got {got}, expected {expected}")

proc vmError(msg: string) =
    error(msg)
    raise newException(VmError, msg)

proc eval(node: Node; scope: Scope): Object =
    result = case node.kind:
        of nkEmpty     : unitObj
        of nkProgram   : evalProgram(node, scope)
        of nkExprParen : evalCall(node, scope)
        of nkLit       : evalLit(node, scope)
        of nkLetStmt   : evalLet(node, scope)
        of nkEqExpr    : evalBlock(node, scope)
        of nkDoExpr    : evalBlock(node, scope)
        of nkId        : evalId(node, scope)
        # of nkFuncStmt   : evalFunc(node, scope)
        # of nkEnumStmt   : evalEnum(node, scope)
        # of nkStructStmt : evalStruct(node, scope)
        # of nkBraceExpr  : evalStructInit(node, scope)
        # of nkBlockStmt  : evalBlock(node, scope)
        # of nkAssign     : evalAssign(node, scope)
        # of nkInfix      : evalInfixOp(node, scope)
        # of nkIfExpr     : evalIfExpr(node, scope)
        # of nkWhileStmt  : evalWhile(node, scope)
        # of nkReturnStmt : evalReturn(node, scope)
        # of nkDotExpr    : evalDotExpr(node, scope)
        # of nkDoExpr     : evalDoExpr(node, scope)
        else: unimplemented(fmt"eval for {node.kind}")

proc evalProgram(program: Node; scope: Scope): Object =
    result = nullObj
    for stmt in program.children:
        result = eval(stmt, scope)


proc evalCall(node: Node; scope: Scope): Object =
    let name = node[0].id
    let args = node[1].children.mapIt(eval(it, scope))

    debug fmt"call '{name}', args: {args}"

    var fnObj = builtins.filterIt(it.id == name)[0]

    if fnObj == nil:
        fnObj = scope[name].Object
    if fnObj == nil:
        vmError(fmt"identifier '{name}' is undefined")
    if fnObj.oType != tyFunc:
        vmError(fmt"identifier '{name}' is not a function")

    let invalid = checkFuncTypes(fnObj.typ, args.mapIt(it.typ))
    if invalid != -1:
        if invalid < fnObj.fnParams.len():
            let paramName = $fnObj.fnParams[invalid].id

            vmError(
                fmt"invalid type for {invalid + 1} parameter " &
                fmt"'{paramName}', " &
                fmt"got {$args[invalid].typ.kind}, " &
                fmt"expected {$fnObj.typ.funcParams[invalid].kind}")
        else:
            vmError(fmt"missing {invalid + 1} argument for '{fnObj.id}', expected {$fnObj.typ.funcParams[invalid].kind}")

    if fnObj.kind == skBuiltinFunc:
        let fn = lookupBuiltinFunc(fnObj)
        result = fn(args)
    elif fnObj.kind == skFunc:
        let fnScope = fnObj.injectArgsIntoScope(args)
        result = evalBlock(fnObj.fnBody, fnScope)
        unwrapReturnVal(result)
    else:
        unreachable()

proc evalLit(node: Node; scope: Scope): Object =
    let lit = node.lit

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
        #[ of tlkEnum:
            var possibleTypes = newSeq[Type]()
            for sym in scope.syms.values():
                if sym.kind == skType and
                   sym.typ.kind == tyEnum and
                   lit.enumVal in sym.typ.enumFields:
                    possibleTypes &= sym.typ

            let resultType = case possibleTypes.len():
                of 1:
                    possibleTypes[0]
                of 0:
                    vmError(fmt"identifier '{lit.enumVal}' is not an enum field")
                else:
                    vmError(fmt"ambiguous identifier '{lit.enumVal}', " &
                            fmt"possible types for it: " &
                            "\n  " & possibleTypes.join("\n  "))
            Object(
                typ     : resultType,
                kind    : skLit,
                oType   : tyEnum,
                enumVal : resultType.enumFields.find(lit.enumVal)) ]#
        # of lkUInt:
        #     if rangeCheck(lit.uintVal, uint32.low..uint32.high):
        #         Object(kind: oU32, u32Val: uint32(lit.uintVal))
        #     else:
        #         Object(kind: oU64, u64Val: uint64(lit.uintVal))
        # of lkFloat:
        #     Object(kind: oF32, f32Val: float32(lit.floatVal))
        of tlkBool: Object(
            typ     : boolType,
            kind    : skLit,
            oType   : tyBool,
            boolVal : lit.boolVal)
        else: unimplemented(fmt"'evalLit' for {lit.kind}")

proc evalFunc(node: Node; scope: Scope): Object =
    let name    = node[0].id
    let body    = node[3]
    var params  = newSeq[Sym]()
    let fnScope = newScopeEnclosed(scope.clone())
    var fnType  = nil.Type

    for param in node[1].children:
        let paramId     = param[0].id
        let paramTypeId = param[1].id
        var paramType   = types.builtinTypes.filterIt(it.name == paramTypeId)[0]

        if paramType == nil:
            let paramTypeSym = scope[paramTypeId]

            if paramTypeSym == nil:
                vmError(fmt"identifier '{paramTypeId}' is undefined")
            if paramTypeSym.kind != skType:
                vmError(fmt"identifier '{paramTypeId}' is not a type")

            paramType = paramTypeSym.typ

        params &= Sym(
            id   : paramId,
            typ  : paramType,
            kind : skParam)

    let returnTypeId = node[2].id
    var returnType   = types.builtinTypes.filterIt(it.name == returnTypeId)[0]

    if returnType == nil:
        let returnTypeSym = scope[returnTypeId]

        if returnTypeSym == nil:
            vmError(fmt"identifier '{returnTypeId}' is undefined")
        if returnTypeSym.kind != skType:
            vmError(fmt"identifier '{returnTypeId}' is not a type")

        returnType = returnTypeSym.typ

    fnType = Type(
        name           : name,
        kind           : tyFunc,
        funcIsVarargs  : false,
        funcParams     : params.mapIt(it.typ),
        funcReturnType : returnType)

    result = Object(
        id       : name,
        typ      : fnType,
        kind     : skFunc,
        oType    : tyFunc,
        fnParams : params,
        fnBody   : body,
        fnScope  : fnScope)
    fnScope[name] = result # for recursion
    scope[name]   = result

#[ proc evalEnum(node: Node; scope: Scope): Object =
    let name   = node.name
    var fields = newSeq[string]()

    for field in node.body.children:
        field.expected(nkEnumField)
        field[0].expected({nkLit, nkId})
        field[1].expected(nkEmpty)

        if field[0].kind == nkLit:
            let fieldLit = field[0].lit

            if fieldLit.kind != tlkEnum:
                vmError(fmt"expected enum literal, got {fieldLit.kind} literal")
            if fields.filterIt(it == fieldLit.enumVal)[0] != "":
                vmError(fmt"duplicate field '{fieldLit.enumVal}'")

            fields &= fieldLit.enumVal
        else:
            let fieldId = field[0].id

            if fields.filterIt(it == fieldId)[0] != "":
                vmError(fmt"duplicate field '{fieldId}'")

            fields &= fieldId

    result = Object(
        id    : name,
        typ   : newEnumType(name, fields),
        kind  : skType,
        oType : tyEnum
    )
    scope[name] = result ]#

proc evalStruct(node: Node; scope: Scope): Object =
    let name   = node[0].id
    var fields = initOrderedTable[string, Type]()

    for field in node[1].children:
        # field.expected(nkColonExpr)
        field[0].expectKind(nkId)
        field[1].expectKind(nkId)

        let fieldName     = field[0].id
        let fieldTypeName = field[1].id
        let fieldTypeKind = toTypeKind(fieldTypeName)

        case fieldTypeKind
        of tyUnknown:
            var fieldTypeObj = scope[fieldTypeName]

            if fieldTypeObj == nil:
                vmError(fmt"identifier '{fieldTypeName}' is undefined")
            if fieldTypeObj.kind != skType:
                vmError(fmt"identifier '{fieldTypeName}' is not a type")

            fields[fieldName] = fieldTypeObj.typ
        of BuiltInTypes + {tyString}:
            fields[fieldName] = getType(fieldTypeKind)
        of tyStruct, tyFunc:
            unreachable()
        else:
            vmError(fmt"type {fieldTypeKind} can't be a field")

    result  = Object(
        id    : name,
        typ   : newStructType(name, fields),
        kind  : skType,
        oType : tyStruct)
    scope[name] = result

proc evalStructInit(node: Node; scope: Scope): Object =
    let typeName   = node[0].id
    let typeObject = scope[typeName].Object

    if typeObject == nil:
        vmError(fmt"identifier '{typeName}' is undefined")
    if typeObject.oType != tyStruct:
        vmError(fmt"identifier '{typeName}' is not a struct")

    var fields = initOrderedTable[string, Object]()

    let fieldExprs =
        if node.len() > 1:
            node.children[1..^1]
        else:
            @[]

    for i, fieldExpr in fieldExprs:
        fieldExpr.expectKind(nkEqExpr)
        fieldExpr[0].expectKind(nkId)

        let fieldId   = fieldExpr[0][0].id
        let fieldExpr = eval(fieldExpr[1], scope)
        let fieldType = typeObject.typ.structFields[fieldId]

        if not isEqTypes(fieldExpr.typ, fieldType):
            errTypeMismatchFor(fmt"field '{fieldId}'", $fieldExpr.typ, $fieldType)

        fields[fieldId] = fieldExpr

    result = Object(
        id           : "",
        kind         : skLit,
        typ          : typeObject.typ,
        oType        : tyStruct,
        structFields : fields
    )

proc evalBlock(node: Node; scope: Scope): Object =
    result = unitObj

    for stmt in node.children:
        result = eval(stmt, scope)
        if result.kind == skReturnVal:
            break

proc evalAssign(node: Node; scope: Scope): Object =
    let name = node[0].id
    var val  = scope[name].Object

    if val == nil:
        vmError(fmt"undeclared variable '{name}'")

    let evaluated = eval(node[1], scope)
    if not isEqTypes(evaluated.typ, val.typ):
        errTypeMismatchFor(fmt"variable '{name}'", $evaluated.oType, $val.oType)

    val[] = Object(
        id    : val.id,
        typ   : val.typ,
        kind  : val.kind,
        oType : val.oType)[]
    val.copyVal(evaluated)
    result = unitObj

proc evalId(node: Node; scope: Scope): Object =
    let name = node.id
    let sym  = block:
        let builtInSyms = builtins.filterIt(it.id == name)

        if builtInSyms.len() < 1: scope[name].Object
        elif builtInSyms.len() > 1: unreachable()
        else: builtInSyms[0]

    if sym == nil:
        vmError(fmt"identifier '{name}' is undefined")

    result = Object(id: name, kind: sym.kind, typ: sym.typ, oType: sym.oType)
    result.copyVal(sym)

# proc evalIfExpr(node: Node; scope: Scope): Object =
#     assert(node.len() > 0, "Illformed AST")
#     result = unitObj

#     for branch in node.children:
#         if branch.kind == nkIfBranch:
#             let cond = eval(branch[0], scope)

#             if not isEqTypes(cond.typ, boolType):
#                 errTypeMismatch($cond.typ, "bool")
#             if cond.boolVal:
#                 return eval(branch.body, scope)
#         else:
#             return eval(branch.body, scope)

proc evalWhile(node: Node; scope: Scope): Object =
    result = unitObj

    while true:
        let cond = eval(node[0], scope)

        if not isEqTypes(cond.typ, boolType):
            errTypeMismatch($cond.typ, "bool")
        if not cond.boolVal: break

        let body = eval(node[1], scope)

        if not isEqTypes(body.typ, unitType):
            errTypeMismatch($cond.typ, "unit")

proc evalReturn(node: Node; scope: Scope): Object =
    result = eval(node[0], scope).wrapReturnValMaybe()

proc evalLet(node: Node; scope: Scope): Object =
    let name = node[0].id
    let typ  = node[1]
    let expr = eval(node[2], scope)

    if typ.kind != nkEmpty:
        let expectedType = block:
            let builtInTypes = builtInTypes.filterIt(it.name == typ.id)

            if builtInTypes.len() < 1: scope[typ.id].typ
            elif builtInTypes.len() > 1: unreachable()
            else: builtInTypes[0]
        if expr.typ != expectedType:
            vmError(fmt"type mismatch, got {expr.typ}, expected {expectedType}")

    if scope[name] != nil:
        vmError(fmt"identifier '{name}' is already defined")

    let variable = Object(
        id    : name,
        typ   : expr.typ,
        kind  : skVar,
        oType : expr.typ.kind
    )
    variable.copyVal(expr)
    scope[name] = variable

    result = unitObj

proc evalDotExpr(node: Node; scope: Scope): Object =
    let left  = node[0]
    let right = node[1]

    result = unitObj

    case left.kind
    of nkId:
        right.expectKind(nkId)

        let leftSym  = scope[left.id]
        let leftType = leftSym.typ

        case leftSym.kind:
        of skType:
            case leftType.kind:
            of tyEnum:
                let fieldName = leftType.enumFields.filterIt(it == right.id)[0]

                if fieldName == "":
                    vmError(fmt"undeclared identifier '{right.id}'")
                else:
                    result = Object(
                        typ     : leftType,
                        kind    : skLit,
                        oType   : tyEnum,
                        enumVal : leftType.enumFields.find(fieldName))
            else: unimplemented(fmt"'evalDotExpr' type {leftType.kind} case")
        of skVar:
            let leftObj = leftSym.Object

            case leftObj.oType
            of tyStruct:
                unimplemented(fmt"field access")
            else:
                vmError(fmt"variable of type {leftObj.oType} has no members")
        else: unimplemented(fmt"'evalDotExpr' {leftSym.kind} case")
    else: unimplemented(fmt"'evalDotExpr' for {left.kind}")

proc evalDoExpr(node: Node; scope: Scope): Object =
    const typeStyle = TextStyle(foreground: BrightCyan, italic: true)
    let evaluated = eval(node[0], scope)
    stdout.writeLine(stylizeText(evaluated.inspect(), typeStyle))
    result = unitObj


# proc evalInfixOp(node: Node; scope: Scope): Object =
#     let op    = node[0].id
#     var left  = eval(node[1], scope)
#     var right = eval(node[2], scope)
#     left.unwrapReturnValMaybe()
#     right.unwrapReturnValMaybe()

#     assert(left != nil)
#     assert(right != nil)

#     result = case op:
#         of "==" : evalEqOp(left, right)
#         of "!=" : evalNeOp(left, right)
#         of "<"  : evalLtOp(left, right)
#         of "<=" : evalLeOp(left, right)
#         of ">"  : evalGtOp(left, right)
#         of ">=" : evalGeOp(left, right)
#         of "+"  : evalSumOp(left, right)
#         of "-"  : evalSubOp(left, right)
#         of "*"  : evalMulOp(left, right)
#         of "/"  : evalDivOp(left, right)
#         of "++" : evalConcatOp(left, right)
#         else: vmError(fmt"invalid operator '{op}'")


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
        else: vmError(fmt"type mismatch, got {param.oType} for builtin_len, expected string")

proc builtin_str(params: seq[Object]): Object =
    if params.len() != 1:
        vmError(fmt"invalid count of arguments, got {params.len()}, expected 1")

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
        vmError(fmt"expected return value, got {obj.kind}")

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
