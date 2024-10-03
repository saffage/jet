package constant

import (
	"fmt"
	"math/big"
	"strconv"
)

//go:generate stringer -type=Kind
type Kind byte

const (
	Int Kind = iota
	Float
	String

	// TODO remove this
	Bool
	Array
)

type Value interface {
	Kind() Kind
}

func NewBigInt(value *big.Int) *intValue {
	if value == nil {
		panic("nil argument")
	}
	return &intValue{value}
}

func NewBigFloat(value *big.Float) *floatValue {
	if value == nil {
		panic("nil argument")
	}
	return &floatValue{value}
}

func NewBool(value bool) *boolValue       { return &boolValue{value} }
func NewInt(value int64) *intValue        { return &intValue{big.NewInt(value)} }
func NewFloat(value float64) *floatValue  { return &floatValue{big.NewFloat(value)} }
func NewString(value string) *stringValue { return &stringValue{value} }
func NewArray(value []Value) *arrayValue  { return &arrayValue{value} }

func IsInt(value Value) bool    { return value.Kind() == Int }
func IsFloat(value Value) bool  { return value.Kind() == Float }
func IsString(value Value) bool { return value.Kind() == String }
func IsBool(value Value) bool   { return value.Kind() == Bool }
func IsArray(value Value) bool  { return value.Kind() == Array }

func AsInt(value Value) *big.Int {
	if IsInt(value) {
		return value.(*intValue).val
	}
	return nil
}

func AsFloat(value Value) *big.Float {
	if IsFloat(value) {
		return value.(*floatValue).val
	}
	return nil
}

func AsString(value Value) *string {
	if IsString(value) {
		return &value.(*stringValue).val
	}
	return nil
}

func AsBool(value Value) *bool {
	if IsBool(value) {
		return &value.(*boolValue).val
	}
	return nil
}

func AsArray(value Value) *[]Value {
	if IsArray(value) {
		return &value.(*arrayValue).val
	}
	return nil
}

//------------------------------------------------
// Value implementation
//------------------------------------------------

type (
	intValue    struct{ val *big.Int }
	floatValue  struct{ val *big.Float }
	stringValue struct{ val string }
	boolValue   struct{ val bool }
	arrayValue  struct{ val []Value }
)

func (v *intValue) String() string    { return v.val.String() }
func (v *floatValue) String() string  { return v.val.String() }
func (v *stringValue) String() string { return strconv.Quote(v.val) }
func (v *boolValue) String() string   { return strconv.FormatBool(v.val) }
func (v *arrayValue) String() string  { return fmt.Sprintf("%v", v.val) }

func (v *intValue) Kind() Kind    { return Int }
func (v *floatValue) Kind() Kind  { return Float }
func (v *stringValue) Kind() Kind { return String }
func (v *boolValue) Kind() Kind   { return Bool }
func (v *arrayValue) Kind() Kind  { return Array }
