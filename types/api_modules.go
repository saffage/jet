package types

import (
	"sync"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
)

// This module contains the declaration of the Jet built-ins.
var ModuleCore *Module

// This module contains C type declarations and other tools for
// interacting with the C backend.
var ModuleC *Module

var onceInitModuleCore sync.Once

func InitModuleCore(cfg *config.Config) {
	_ = cfg
	onceInitModuleCore.Do(initModuleCore)
}

func initModuleCore() {
	var (
		IntNode    = &ast.TypeDecl{Name: &ast.Upper{Data: "Int"}}
		FloatNode  = &ast.TypeDecl{Name: &ast.Upper{Data: "Float"}}
		StringNode = &ast.TypeDecl{Name: &ast.Upper{Data: "String"}}
		NoneNode   = &ast.TypeDecl{Name: &ast.Upper{Data: "None"}, Expr: &ast.Block{}}
	)

	ModuleCore = NewModule(NewNamedEnv(nil, "core"), "core", ast.File{
		Ast: &ast.Stmts{
			Nodes: []ast.Node{IntNode, FloatNode, StringNode, NoneNode},
		},
	})

	var (
		IntDef    = NewTypeDef(ModuleCore.Env, NewTypeDesc(IntType), IntNode)
		FloatDef  = NewTypeDef(ModuleCore.Env, NewTypeDesc(FloatType), FloatNode)
		StringDef = NewTypeDef(ModuleCore.Env, NewTypeDesc(StringType), StringNode)
		NoneDef   = NewTypeDef(ModuleCore.Env, NewTypeDesc(NoneType), NoneNode)
	)

	ModuleCore.Env.types = make(map[string]*TypeDef)
	ModuleCore.Env.types["Int"] = IntDef
	ModuleCore.Env.types["Float"] = FloatDef
	ModuleCore.Env.types["String"] = StringDef
	ModuleCore.Env.types["None"] = NoneDef

	var NoneVariant = newConstructor(
		ModuleCore.Env,
		NoneDef,
		&Value{T: NoneDef.t},
		&ast.Variant{Name: &ast.Upper{Data: "None"}},
	)

	ModuleCore.Env.symbols = make(map[string]Symbol)
	ModuleCore.Env.symbols["None"] = NoneVariant
}
