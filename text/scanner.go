package text

import (
	"slices"
	"strings"
)

type Scanner struct {
	fileID FileID // Needed for token position.
	buf    []byte // Actual data.
	bufPos int    // Current character index.
}

func NewScanner(buffer []byte, id FileID) *Scanner {
	return &Scanner{
		fileID: id,
		buf:    buffer,
	}
}

// Returns the current character.
func (base *Scanner) Peek() byte {
	return base.LookAhead(0)
}

// Returns the previous character. Maybe can panic.
func (base *Scanner) Prev() byte {
	return base.LookAhead(-1)
}

// Returns character with specified offset.
func (base *Scanner) LookAhead(offset int) byte {
	if base.bufPos+offset < len(base.buf) {
		return base.buf[base.bufPos+offset]
	}

	return '\000'
}

// Returns the current character and advances forward.
func (base *Scanner) Advance() (previous byte) {
	previous = base.Peek()

	switch previous {
	case '\000':
		// Stay here

	case '\n', '\r':
		if base.Peek() == '\r' {
			base.bufPos++
		}

		if base.Peek() == '\n' {
			base.bufPos++
		}

	default:
		base.bufPos++
	}

	return
}

// Consumes any of `chars` and returns true, otherwise returns false.
func (base *Scanner) Consume(chars ...byte) bool {
	if len(chars) == 0 || base.Match(chars...) {
		base.Advance()
		return true
	}

	return false
}

// Returns true if the current character matches `char`.
func (base *Scanner) Match(chars ...byte) bool {
	return slices.Contains(chars, base.Peek())
}

// Takes all `data` while not `stop`.
//
// Note: `function` should advance on every iteration.
func (base *Scanner) Take(f func() (data []byte, stop bool)) string {
	result := strings.Builder{}

	for base.bufPos < len(base.buf) {
		data, stop := f()
		result.Write(data)

		if stop {
			break
		}
	}

	return result.String()
}

// Takes all characters while `predicate` returns true.
func (s *Scanner) TakeWhile(predicate func(byte) bool) string {
	return s.Take(func() ([]byte, bool) {
		if predicate(s.Peek()) {
			return []byte{s.Advance()}, false
		}
		return nil, true
	})
}

// Takes all characters while `predicate` returns false.
func (base *Scanner) TakeUntil(predicate func(byte) bool) string {
	return base.Take(func() ([]byte, bool) {
		if !predicate(base.Peek()) {
			return []byte{base.Advance()}, false
		}
		return nil, true
	})
}

func (base *Scanner) Pos() Pos {
	return PosFrom(base.fileID, base.bufPos)
}

func (base *Scanner) PrevPos() Pos {
	return PosFrom(base.fileID, base.bufPos-1)
}
