package types

import "testing"

func TestEquals(t *testing.T) {
	x := NewStruct(map[string]Type{
		"age":   Primitives[I32],
		"adult": Primitives[Bool],
	})
	y := NewStruct(map[string]Type{
		"age":   Primitives[I32],
		"adult": Primitives[Bool],
	})

	if !x.Equals(x) {
		t.Errorf("struct types are not equals, but should:\nx: %s\ny: %s", x, y)
	}

	if !x.Equals(y) {
		t.Errorf("struct types are not equals, but should:\nx: %s\ny: %s", x, y)
	}
}
