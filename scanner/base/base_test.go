package base

import (
	"fmt"
	"testing"
)

func TestTake(t *testing.T) {
	buffer := "001"
	s := New(([]byte)(buffer), 0)

	data := s.Take(func() (data []byte, stop bool) {
		fmt.Printf("pos: %d; byte: %c\n", s.bufPos, s.Peek())
		if s.Peek() == '0' {
			return []byte{s.Advance()}, false
		}
		return nil, true
	})

	if data != "00" {
		t.Errorf("expected '%s'; got '%s'", "00", data)
	}
}
