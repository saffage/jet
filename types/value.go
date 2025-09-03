package types

import (
	"fmt"
	"math/big"
	"strconv"
)

// Represents a compile-time known value.
// Also can represent just a type in some cases.
type Value struct {
	T Type
	V ConstantValue
}

//go:generate go tool stringer -type=Kind -output=value_string.go
type Kind byte

const (
	ConstantInt    Kind = iota // int constant
	ConstantFloat              // float constant
	ConstantString             // string constant
	ConstantBool               // bool constant
	ConstantArray              // array constant
)

type ConstantValue interface {
	Kind() Kind
	implValue()
}

func NewBigIntConstant(value *big.Int) ConstantValue {
	if value == nil {
		panic("nil argument")
	}
	return &intValue{value}
}

func NewBigFloatConstant(value *big.Float) ConstantValue {
	if value == nil {
		panic("nil argument")
	}
	return &floatValue{value}
}

func NewBoolConstant(value bool) ConstantValue             { return &boolValue{value} }
func NewIntConstant(value int64) ConstantValue             { return &intValue{big.NewInt(value)} }
func NewFloatConstant(value float64) ConstantValue         { return &floatValue{big.NewFloat(value)} }
func NewStringConstant(value string) ConstantValue         { return &stringValue{value} }
func NewArrayConstant(value []ConstantValue) ConstantValue { return &arrayValue{value} }

func IsIntConstant(value ConstantValue) bool    { return value.Kind() == ConstantInt }
func IsFloatConstant(value ConstantValue) bool  { return value.Kind() == ConstantFloat }
func IsStringConstant(value ConstantValue) bool { return value.Kind() == ConstantString }
func IsBoolConstant(value ConstantValue) bool   { return value.Kind() == ConstantBool }
func IsArrayConstant(value ConstantValue) bool  { return value.Kind() == ConstantArray }

func AsIntConstant(value ConstantValue) *big.Int {
	if IsIntConstant(value) {
		return value.(*intValue).val
	}
	return nil
}

func AsFloatConstant(value ConstantValue) *big.Float {
	if IsFloatConstant(value) {
		return value.(*floatValue).val
	}
	return nil
}

func AsStringConstant(value ConstantValue) *string {
	if IsStringConstant(value) {
		val := value.(*stringValue).val
		return &val
	}
	return nil
}

func AsBoolConstant(value ConstantValue) *bool {
	if IsBoolConstant(value) {
		val := value.(*boolValue).val
		return &val
	}
	return nil
}

func AsArrayConstant(value ConstantValue) *[]ConstantValue {
	if IsArrayConstant(value) {
		val := value.(*arrayValue).val
		return &val
	}
	return nil
}

//------------------------------------------------
// ConstantValue implementation
//------------------------------------------------

type (
	intValue    struct{ val *big.Int }
	floatValue  struct{ val *big.Float }
	stringValue struct{ val string }
	boolValue   struct{ val bool }
	arrayValue  struct{ val []ConstantValue }
)

func (v *intValue) String() string    { return v.val.String() }
func (v *floatValue) String() string  { return v.val.String() }
func (v *stringValue) String() string { return strconv.Quote(v.val) }
func (v *boolValue) String() string   { return strconv.FormatBool(v.val) }
func (v *arrayValue) String() string  { return fmt.Sprintf("%v", v.val) }

func (v *intValue) Kind() Kind    { return ConstantInt }
func (v *floatValue) Kind() Kind  { return ConstantFloat }
func (v *stringValue) Kind() Kind { return ConstantString }
func (v *boolValue) Kind() Kind   { return ConstantBool }
func (v *arrayValue) Kind() Kind  { return ConstantArray }

func (intValue) implValue()    {}
func (floatValue) implValue()  {}
func (stringValue) implValue() {}
func (boolValue) implValue()   {}
func (arrayValue) implValue()  {}
