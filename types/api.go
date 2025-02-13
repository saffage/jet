package types

import (
	"errors"
	"sync"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

// TODO introduce module hierarchy.

func Check(f *text.File, stmts *ast.Stmts) (*Module, error) {
	InitModuleCore()

	var (
		moduleName = f.Name
		scope      = NewNamedEnv(ModuleEnv, nil, moduleName)
		module     = NewModule(scope, moduleName, f, stmts)
		check      = &checker{module: module, env: scope}
	)

	_ = scope.Use(ModuleCore.Env)

	report.DebugX("types", "checking module '%s'", moduleName)

loop:
	for _, node := range stmts.Items {
		switch node := node.(type) {
		case *ast.LetDecl:
			check.resolveLetDecl(node)

		case *ast.TypeAlias:
			check.resolveTypeAlias(node)

		case *ast.TypeDef:
			check.resolveTypeDef(node)

		default:
			if _, err := check.eval(node); err != nil {
				break loop
			}
		}
	}

	module.completed = true
	return check.module, errors.Join(check.errors...)
}

func CheckFile(f *text.File) (*Module, error) {
	stmts, err := parser.ParseFile(f, token.DefaultFlags, parser.DefaultFlags)

	if err != nil {
		return nil, err
	}

	return Check(f, stmts)
}

// This module contains the declaration of the Jet built-ins.
var ModuleCore *Module

// This module contains C type declarations and other tools for
// interacting with the C backend.
var ModuleC *Module

var onceInitModuleCore sync.Once

func InitModuleCore() {
	onceInitModuleCore.Do(func() {
		ModuleCore = NewModule(
			NewNamedEnv(ModuleEnv, nil, "core"),
			"core",
			nil,
			nil,
		)

		var (
			NoneTypeAlias   = NewExternTypeAlias(ModuleCore.Env, NoneType)
			NeverTypeAlias  = NewExternTypeAlias(ModuleCore.Env, NeverType)
			IntTypeAlias    = NewExternTypeAlias(ModuleCore.Env, IntType)
			FloatTypeAlias  = NewExternTypeAlias(ModuleCore.Env, FloatType)
			StringTypeAlias = NewExternTypeAlias(ModuleCore.Env, StringType)
			BoolTypeAlias   = NewExternTypeDef(ModuleCore.Env, nil, BoolType)

			TrueVariant  = NewVariant(ModuleCore.Env, nil, BoolType.Variant(0), nil, nil)
			FalseVariant = NewVariant(ModuleCore.Env, nil, BoolType.Variant(1), nil, nil)
		)

		ModuleCore.Env.symbols = map[string]Symbol{
			"None":   NoneTypeAlias,
			"Never":  NeverTypeAlias,
			"Int":    IntTypeAlias,
			"Float":  FloatTypeAlias,
			"String": StringTypeAlias,
			"Bool":   BoolTypeAlias,
			"True":   TrueVariant,
			"False":  FalseVariant,
		}
	})
}
