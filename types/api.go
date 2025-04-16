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

func Check(f *text.File, stmts *ast.Stmts) (*Module, error) {
	InitPrelude()

	var (
		moduleName = f.Name
		scope      = NewNamedEnv(ModuleEnv, nil, moduleName)
		module     = NewModule(scope, moduleName, f, stmts)
		check      = &checker{module: module, env: scope}
	)

	_ = scope.Use(prelude.Env)

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
	return check.module, report.Join(check.errors...)
}

func CheckFile(f *text.File) (*Module, error) {
	stmts, err := parser.
		FromFile(f, token.DefaultFlags, parser.DefaultFlags, nil).
		ParseOrError()

	if err != nil {
		return nil, err
	}

	return Check(f, stmts)
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
