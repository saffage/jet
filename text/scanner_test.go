package text

import "testing"

func TestChars(t *testing.T) {
	buffer := "001"
	s := NewScanner(([]byte)(buffer), 0)

	data := s.TakeWhile(func(char rune) bool { return char == '0' })

	if data != "00" {
		t.Errorf("expected '%s'; got '%s'", "00", data)
	}
}
