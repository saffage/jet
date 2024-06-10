package types

import "testing"

func TestNamed(t *testing.T) {
	ty := Unit
	x := NewNamed("X", ty)
	y := NewNamed("Y", ty)
	test(t, x, y, false)
	test(t, x.Underlying(), y.Underlying(), true)

	y = x
	test(t, x, y, true)
	test(t, x.Underlying(), y.Underlying(), true)
}

func test(t *testing.T, x, y Type, isEquals bool) {
	if x.Equals(y) {
		if !isEquals {
			t.Errorf("expected types '%s' and '%s' to be not equals", x, y)
		}
	} else if isEquals {
		t.Errorf("expected types '%s' and '%s' to be equals", x, y)
	}
}
