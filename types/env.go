package types

import (
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

type Env struct {
	symbols  map[string]Symbol
	children []*Env
	parent   *Env
	parentID int
	name     string // for debugging
	kind     EnvKind
}

//go:generate stringer -type=EnvKind -linecomment -output=env_kind_string.go
type EnvKind byte

const (
	_ EnvKind = iota

	ModuleEnv   // module scope
	ScopeEnv    // block scope
	TypeEnv     // type body
	FnEnv       // function body
	FnParamsEnv // function parameters
)

func NewEnv(kind EnvKind, parent *Env) *Env {
	return NewNamedEnv(kind, parent, "")
}

func NewNamedEnv(kind EnvKind, parent *Env, name string) *Env {
	if kind != ModuleEnv {
		assert(parent != nil)
	}

	env := &Env{kind: kind, parent: parent, name: name}

	if parent != nil && parent.parent != nil {
		env.parentID = len(parent.children)
		parent.children = append(parent.children, env)
	}

	return env
}

func (env *Env) Kind() EnvKind {
	return env.kind
}

func (env *Env) Parent() *Env {
	return env.parent
}

func (env *Env) ParentID() int {
	return env.parentID
}

func (env *Env) Children() []*Env {
	return env.children
}

func (env *Env) Child(n int) *Env {
	return env.children[n]
}

// Defines a new symbol in the environment. If a symbol with the same
// name is already defined in this environment, it will return it without
// defining a new symbol.
func (env *Env) Define(symbol Symbol) Symbol {
	if symbol == nil {
		// Environment should not contain a nil symbols.
		panic("attempt to define nil symbol")
	}

	if symbol.Name() == "_" {
		return nil
	}

	if defined := env.LookupLocal(symbol.Name()); defined != nil {
		return defined
	}

	if env.symbols == nil {
		env.symbols = make(map[string]Symbol)
	}

	env.symbols[symbol.Name()] = symbol
	report.Debug(
		"defined symbol `%s` of type `%s` in '%s' %s",
		symbol.Name(),
		symbol.Type(),
		env.Path(),
		env.kind.String(),
	)
	return nil
}

// Redefines an existing symbol and returns it, or defined a new symbol.
func (env *Env) Redefine(symbol Symbol) Symbol {
	if defined := env.Define(symbol); defined != nil {
		env.symbols[symbol.Name()] = symbol
		return defined
	}

	return nil
}

// Searches for the specified symbol by name in the context of
// the specified environment and returns it, or nil if such symbol
// is undefined or unavailable (private).
func (env *Env) Lookup(name string) (Symbol, *Env) {
	if member := env.LookupLocal(name); member != nil {
		return member, env
	}

	if env.parent != nil {
		return env.parent.Lookup(name)
	}

	return nil, nil
}

// Searches for the specified symbol by name in the specified
// environment and returns it, or nil if no such symbol is defined
// in the current environment.
func (env *Env) LookupLocal(name string) Symbol {
	if env.symbols == nil {
		return nil
	}

	return env.symbols[name]
}

func (env *Env) Use(other *Env, names ...ast.Ident) error {
	if len(names) == 0 {
		for _, symbol := range other.symbols {
			env.Redefine(symbol)
		}

		return nil
	}

	var errs []error

	for _, name := range names {
		if symbol := other.LookupLocal(name.Name()); symbol != nil {
			env.Redefine(symbol)
		} else {
			errs = append(errs, errUndefinedIdent(name, nil))
		}
	}

	return report.Join(errs...)
}

func (env *Env) Path() string {
	namesBeforeThisEnv := []string(nil)

	for env := env.parent; env != nil; env = env.parent {
		// FIXME: path is incorrect, currently used only for debugging.

		if env.name == "" {
			namesBeforeThisEnv = append(namesBeforeThisEnv, "#["+env.kind.String()+"]#")
		} else {
			namesBeforeThisEnv = append(namesBeforeThisEnv, env.name)
		}
	}

	buf := strings.Builder{}

	for i := len(namesBeforeThisEnv) - 1; i >= 0; i-- {
		buf.WriteString(namesBeforeThisEnv[i])
		buf.WriteByte('/')
	}

	buf.WriteString(env.name)
	return buf.String()
}
