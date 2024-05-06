package constant

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/saffage/jet/ast"
)

type Value interface {
	Kind() Kind
	String() string

	implValue()
}

func NewBool(value bool) Value {
	return boolValue{value}
}

func NewInt(value *big.Int) Value {
	return intValue{value}
}

func NewFloat(value *big.Float) floatValue {
	return floatValue{value}
}

func NewString(value string) Value {
	return stringValue{value}
}

func FromNode(node *ast.Literal) Value {
	if node == nil {
		panic("unnreachable")
	}

	switch node.Kind {
	case ast.IntLiteral:
		if value, ok := new(big.Int).SetString(node.Value, 0); ok {
			return NewInt(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid integer value for constant: '%s'", node.Value))

	case ast.FloatLiteral:
		if value, ok := new(big.Float).SetString(node.Value); ok {
			return NewFloat(value)
		}

		// Unreachable?
		panic(fmt.Sprintf("invalid float value for constant: '%s'", node.Value))

	case ast.StringLiteral:
		return NewString(node.Value)

	default:
		panic("unreachable")
	}
}

func AsBool(value Value) *bool {
	if v, ok := value.(boolValue); ok {
		return &v.bool
	}

	return nil
}

func AsInt(value Value) *big.Int {
	if v, ok := value.(intValue); ok && v.Int != nil {
		return v.Int
	}

	return nil
}

func AsFloat(value Value) *big.Float {
	if v, ok := value.(floatValue); ok && v.Float != nil {
		return v.Float
	}

	return nil
}

func AsString(value Value) *string {
	if v, ok := value.(stringValue); ok {
		return &v.string
	}

	return nil
}

type Kind byte

const (
	Unknown Kind = iota

	Bool // TODO delete and implement through attributes.
	Int
	Float
	String
)

type (
	boolValue   struct{ bool }
	intValue    struct{ *big.Int }
	floatValue  struct{ *big.Float }
	stringValue struct{ string }
)

func (boolValue) implValue()   {}
func (intValue) implValue()    {}
func (floatValue) implValue()  {}
func (stringValue) implValue() {}

func (v boolValue) Kind() Kind     { return Bool }
func (v boolValue) String() string { return strconv.FormatBool(v.bool) }

func (v intValue) Kind() Kind     { return Int }
func (v intValue) String() string { return v.Int.String() }

func (v floatValue) Kind() Kind     { return Float }
func (v floatValue) String() string { return v.Float.String() }

func (v stringValue) Kind() Kind     { return String }
func (v stringValue) String() string { return "\"" + v.string + "\"" }
