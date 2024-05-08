package types

import "testing"

func TestUnit(t *testing.T) {
	if !Unit.Equals(NewTuple()) {
		t.Error("[types.Unit] must be equals to empty tuple")
	}

	if Unit.Underlying() != Unit {
		t.Error("underlying type of the empty tuple must be equal to the tuple itself")
	}
}

func TestSingleElem(t *testing.T) {
	elem0 := I32
	tuple := NewTuple(elem0)

	if !tuple.Equals(elem0) {
		t.Error("single element tuple type must be equals to element type")
	}

	if tuple.Underlying() != elem0 {
		t.Error("single element tuple type must be the same with element type")
	}
}

func TestMultiElem(t *testing.T) {
	elem0 := I32
	elem1 := Bool
	tuple := NewTuple(elem0, elem1)

	if tuple.Equals(elem0) || tuple.Equals(elem1) {
		t.Error("tuple must not be equal to any element type")
	}

	if tuple.Underlying() != tuple {
		t.Error("tuple's underlying type must be equal to the tuple itself")
	}
}
