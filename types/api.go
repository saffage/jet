package types

import (
	"sync"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

// TODO introduce module hierarchy.

func Check(f *text.File, stmts *ast.Stmts, errorHandler func(error)) *Module {
	InitPrelude()

	var (
		moduleName = f.Name
		env        = NewNamedEnv(ModuleEnv, nil, moduleName)
		module     = NewModule(env, moduleName, f, stmts)
		resolver   = &resolver{
			currentModule: module,
			currentEnv:    env,
			errorHandler:  errorHandler,
		}
	)

	_ = env.Use(prelude.Env)

	report.DebugX("types", "checking module '%s'", moduleName)
	ast.WalkTopDown(stmts, resolver)
	module.completed = true
	return module
}

func CheckFile(f *text.File, errorHandler func(error)) *Module {
	stmts := parser.
		FromFile(f, token.DefaultFlags, parser.DefaultFlags, errorHandler).
		Parse()

	return Check(f, stmts, errorHandler)
}

var (
	preludeSync sync.Once
	prelude     *Module
)

func InitPrelude() {
	preludeSync.Do(func() {
		prelude = NewModule(
			NewNamedEnv(ModuleEnv, nil, "core"),
			"core",
			nil,
			nil,
		)

		var (
			UnitTypeAlias   = NewExternalTypeAlias(prelude.Env, UnitType)
			NeverTypeAlias  = NewExternalTypeAlias(prelude.Env, NeverType)
			IntTypeAlias    = NewExternalTypeAlias(prelude.Env, IntType)
			FloatTypeAlias  = NewExternalTypeAlias(prelude.Env, FloatType)
			StringTypeAlias = NewExternalTypeAlias(prelude.Env, StringType)
			BoolTypeAlias   = NewExternalTypeDef(prelude.Env, nil, BoolType)

			TrueVariant  = NewVariant(prelude.Env, nil, BoolType.Variant(0), nil, nil)
			FalseVariant = NewVariant(prelude.Env, nil, BoolType.Variant(1), nil, nil)
		)

		prelude.Env.symbols = map[string]Symbol{
			"Unit":   UnitTypeAlias,
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
