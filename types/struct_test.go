package types

import "testing"

func TestEquals(t *testing.T) {
	x := NewStruct(StructField{"age", I32}, StructField{"adult", Bool})
	y := NewStruct(StructField{"age", I32}, StructField{"adult", Bool})
	z := NewStruct(StructField{"adult", Bool}, StructField{"age", I32})

	if !x.Equals(x) {
		t.Errorf("struct types are not equals, but should:\nx: %s\ny: %s", x, y)
	}

	if !x.Equals(y) {
		t.Errorf("struct types are not equals, but should:\nx: %s\ny: %s", x, y)
	}

	if x.Equals(z) {
		t.Errorf("struct types are equals, but shouldn't:\nx: %s\ny: %s", x, z)
	}
}
