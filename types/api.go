package types

import (
	"sync"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

// TODO introduce module hierarchy.

type CheckerOptions struct {
	parser.Options
}

func Check(f *text.File, stmts *ast.Stmts, opts ...CheckerOptions) *Module {
	InitPrelude()

	opt := CheckerOptions{}

	if len(opts) != 0 {
		opt = opts[0]
	}

	var (
		moduleName = f.Name
		env        = NewNamedEnv(ModuleEnv, nil, moduleName)
		module     = NewModule(env, moduleName, f, stmts)
		// resolver   = &resolver{
		// 	currentModule: module,
		// 	currentEnv:    env,
		// 	errorHandler:  opt.ErrorHandler,
		// }
	)

	_ = env.Use(prelude.Env)
	_ = opt

	report.DebugX("types", "checking module '%s'", moduleName)

	// ast.WalkTopDown(stmts, resolver)

	module.completed = true
	return module
}

func CheckFile(f *text.File, opts ...CheckerOptions) *Module {
	opt := CheckerOptions{}

	if len(opts) != 0 {
		opt = opts[0]
	}

	stmts := parser.
		FromFile(f, opt.Options).
		Parse()

	return Check(f, stmts, opt)
}

// TODO compiler-builtin prelude module must be replaced by implicit module
// import with external declaration.

var (
	preludeOnce sync.Once
	prelude     *Module
)

func InitPrelude() {
	preludeOnce.Do(initPrelude)
}

func initPrelude() {
	prelude = NewModule(
		NewNamedEnv(ModuleEnv, nil, "core"),
		"core",
		nil,
		nil,
	)

	/* var (
		UnitTypeAlias   = NewExternalTypeAlias(prelude.Env, UnitType)
		NeverTypeAlias  = NewExternalTypeAlias(prelude.Env, NeverType)
		IntTypeAlias    = NewExternalTypeAlias(prelude.Env, IntType)
		FloatTypeAlias  = NewExternalTypeAlias(prelude.Env, FloatType)
		StringTypeAlias = NewExternalTypeAlias(prelude.Env, StringType)
		BoolTypeAlias   = NewExternalTypeDef(prelude.Env, nil, BoolType)

		FalseVariant = NewVariant(prelude.Env, nil, BoolType.Variant(0), nil, nil)
		TrueVariant  = NewVariant(prelude.Env, nil, BoolType.Variant(1), nil, nil)
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
	} */
}
