package types

import "github.com/saffage/jet/ast"

type Module struct {
	*TypeInfo

	Env     *Env
	Imports []*Module

	file      ast.File
	name      string
	completed bool
}

func NewModule(env *Env, name string, file ast.File) *Module {
	return &Module{
		TypeInfo: newTypeInfo(),
		Env:      env,
		file:     file,
		name:     name,
	}
}

func (m *Module) Type() Type       { return moduleType{} }
func (m *Module) Name() string     { return m.name }
func (m *Module) Node() ast.Node   { return m.file.Ast }
func (m *Module) Ident() ast.Ident { return nil } // TODO: just use *ast.Name with zero-initialized range
func (m *Module) Owner() *Env      { return m.Env.parent }

func (m *Module) TypeOf(expr ast.Node) Type {
	if expr != nil {
		if t := m.TypeInfo.TypeOf(expr); t != nil {
			return t
		}
		if ident, _ := expr.(*ast.Lower); ident != nil {
			if sym := m.SymbolOf(ident); sym != nil {
				return sym.Type()
			}
		}
	}
	return nil
}

func (m *Module) ValueOf(expr ast.Node) *Value {
	if expr != nil {
		if t := m.TypeInfo.ValueOf(expr); t != nil {
			return t
		}
		// if ident, _ := expr.(*ast.Name); ident != nil {
		// 	if _const, _ := m.SymbolOf(ident).(*Const); _const != nil {
		// 		return _const.value
		// 	}
		// }
	}
	return nil
}

func (m *Module) SymbolOf(ident ast.Ident) Symbol {
	if sym := m.TypeInfo.SymbolOf(ident); sym != nil {
		return sym
	}
	if sym, _ := m.Env.Lookup(ident.String()); sym != nil {
		return sym
	}
	return nil
}

func (m *Module) TypeSymbolOf(ident ast.Ident) *TypeDef {
	if sym := m.TypeInfo.TypeSymbolOf(ident); sym != nil {
		return sym
	}
	if sym, _ := m.Env.LookupType(ident.String()); sym != nil {
		return sym
	}
	return nil
}
