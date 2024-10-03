package types

import (
	"errors"

	"github.com/saffage/jet/report"
)

type Env struct {
	parent   *Env
	symbols  map[string]Symbol
	types    map[string]*TypeDef
	name     string
	children []*Env
	parentID int
}

func NewEnv(parent *Env) *Env {
	return NewNamedEnv(parent, "")
}

func NewNamedEnv(parent *Env, name string) *Env {
	env := &Env{name: name, parent: parent}

	if parent != nil && parent.parent != nil {
		env.parentID = len(parent.children)
		parent.children = append(parent.children, env)
	}

	return env
}

// Returns a name of the environment. Used in code generator. Name can be next:
//
//   - module <name>
//   - func <name>
//   - struct <name>
//   - enum <name>
//   - block
func (env *Env) Name() string {
	return env.name
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
// name is already defined in this environment, it will return error.
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
	report.DebugX("env", "defined symbol `%s` in the env \"%s\"", symbol.Name(), env.name)
	return nil
}

// Defines a new type in the environment. If a type with the same
// name is already defined in this environment, it will return this type.
func (env *Env) DefineType(t *TypeDef) *TypeDef {
	if t == nil {
		// Environment should not contain a nil types.
		panic("attempt to define nil type")
	}

	if t.Name() == "_" {
		return nil
	}

	if defined := env.LookupLocalType(t.Name()); defined != nil {
		return defined
	}

	if env.types == nil {
		env.types = make(map[string]*TypeDef)
	}

	env.types[t.Name()] = t
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

// Searches for the specified type by name in the context of
// the specified environment and returns it, or nil if such type
// is undefined or unavailable (private).
func (env *Env) LookupType(name string) (*TypeDef, *Env) {
	if member := env.LookupLocalType(name); member != nil {
		return member, env
	}

	if env.parent != nil {
		return env.parent.LookupType(name)
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

// Searches for the specified type by name in the specified
// environment and returns it, or nil if no such type is defined
// in the current environment.
func (env *Env) LookupLocalType(name string) *TypeDef {
	if env.types == nil {
		return nil
	}

	return env.types[name]
}

func (env *Env) Use(other *Env, symbols ...string) error {
	var errs []error

	if env.symbols == nil {
		env.symbols = make(map[string]Symbol)
	}

	if len(symbols) == 0 {
		for _, symbol := range other.symbols {
			name := symbol.Name()
			s, defined := env.symbols[name]
			if defined {
				_ = s
				panic("todo")
			}
			env.symbols[name] = symbol
		}
	} else {
		for _, name := range symbols {
			s, defined := env.symbols[name]
			if defined {
				_ = s
				panic("todo")
			}
			s, defined = other.symbols[name]
			if !defined {
				panic("todo")
			}
			env.symbols[name] = s
		}
	}

	return errors.Join(errs...)
}

func (env *Env) UseTypes(other *Env, types ...string) error {
	var errs []error

	if env.types == nil {
		env.types = make(map[string]*TypeDef)
	}

	if len(types) == 0 {
		for _, t := range other.types {
			name := t.Name()
			s, defined := env.types[name]
			if defined {
				_ = s
				panic("todo")
			}
			env.types[name] = t
		}
	} else {
		for _, name := range types {
			s, defined := env.types[name]
			if defined {
				_ = s
				panic("todo")
			}
			s, defined = other.types[name]
			if !defined {
				panic("todo")
			}
			env.types[name] = s
		}
	}

	return errors.Join(errs...)
}
