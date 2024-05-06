package checker

import (
	"github.com/saffage/jet/constant"
	"github.com/saffage/jet/types"
)

// Represents a compile-time known value.
// Also can represent a type in some situations.
type Value struct {
	Type  types.Type
	Value constant.Value // Can be nil.
}
