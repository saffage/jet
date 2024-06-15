package cgen

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

func (gen *generator) globalVarDecl(sym *checker.Var) {
	t := gen.TypeString(sym.Type())
	gen.declVarsSect.WriteString(fmt.Sprintf("%s %s;\n", t, gen.name(sym)))
}

func (gen *generator) varDecl(sym *checker.Var) string {
	return fmt.Sprintf("%s %s;\n", gen.TypeString(sym.Type()), gen.name(sym))
}

func (gen *generator) tempVar(ty types.Type) *checker.Var {
	id := fmt.Sprintf("tmp__%d", gen.funcTempVarId)
	decl := &ast.Decl{Ident: &ast.Ident{Name: id}}
	sym := checker.NewVar(gen.scope, ty, decl)
	gen.line(gen.varDecl(sym))
	_ = gen.scope.Define(sym)
	gen.funcTempVarId++
	return sym
}

func (gen *generator) resultVar(ty types.Type) *checker.Var {
	if ty == nil {
		return nil
	}
	decl := &ast.Decl{Ident: &ast.Ident{Name: "__result"}}
	sym := checker.NewVar(gen.scope, ty, decl)
	if !types.IsArray(ty) {
		gen.linef("%s __result;\n", gen.TypeString(ty))
	}
	_ = gen.scope.Define(sym)
	return sym
}

func (gen *generator) initFunc() {
	gen.linef("void init%s(void)\n{\n", gen.Module.Name())
	gen.indent++

	for def := gen.Defs.Front(); def != nil; def = def.Next() {
		def := def.Value

		if _var, _ := def.(*checker.Var); _var != nil && _var.IsGlobal() && _var.Value() != nil {
			gen.linef("%s;\n", gen.binary(
				_var.Node().(*ast.Decl).Ident,
				_var.Value(),
				types.Unit,
				ast.OperatorAssign,
			))
		}
	}

	gen.indent--
	gen.line("}\n")
}
