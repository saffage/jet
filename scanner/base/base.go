package base

import (
	"math"
	"slices"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/token"
)

type Base struct {
	fileID          config.FileID // Needed for token position.
	buf             []byte        // Actual data.
	bufPos          int           // Current character index.
	lineNum         uint32        // Current line number.
	charNum         uint32        // Current character number.
	prevLineCharNum uint32        // Last character number in the previous line (needed for [PrevPos] function).
}

func New(buffer []byte, fileID config.FileID) *Base {
	return &Base{
		fileID:  fileID,
		buf:     buffer,
		lineNum: 1,
		charNum: 1,
	}
}

// Returns the current character.
func (base *Base) Peek() byte {
	return base.LookAhead(0)
}

// Returns the previous character. Maybe can panic.
func (base *Base) Prev() byte {
	return base.LookAhead(-1)
}

// Returns character with specified offset.
func (base *Base) LookAhead(offset int) byte {
	if base.bufPos+offset < len(base.buf) {
		return base.buf[base.bufPos+offset]
	}

	return '\000'
}

// Returns the current character and advances forward.
func (base *Base) Advance() (previous byte) {
	previous = base.Peek()

	switch previous {
	case '\000':
		// Stay here

	case '\n', '\r':
		base.HandleNewline()

	default:
		base.bufPos++
		base.charNum++
	}

	return
}

// Comsumes any of `chars` and returns true, otherwise returns false.
func (base *Base) Consume(chars ...byte) bool {
	if len(chars) == 0 || base.Match(chars...) {
		base.Advance()
		return true
	}

	return false
}

// Returns true if the current character matches `char`.
func (base *Base) Match(chars ...byte) bool {
	return slices.Contains(chars, base.Peek())
}

// Takes all `data` while not `stop`.
//
// Note: `function` should advance on every iteration.
func (base *Base) Take(f func() (data []byte, stop bool)) string {
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
func (base *Base) TakeWhile(predicate func(byte) bool) string {
	return base.Take(func() ([]byte, bool) {
		if predicate(base.Peek()) {
			return []byte{base.Advance()}, false
		}
		return nil, true
	})
}

// Takes all characters while `predicate` returns false.
func (base *Base) TakeUntil(predicate func(byte) bool) string {
	return base.Take(func() ([]byte, bool) {
		if !predicate(base.Peek()) {
			return []byte{base.Advance()}, false
		}
		return nil, true
	})
}

// Handles a new line character to update internal data.
// Must be called for every line in the buffer.
func (base *Base) HandleNewline() (wasNewline bool) {
	if base.Peek() == '\r' {
		base.bufPos++
		wasNewline = true
	}

	if base.Peek() == '\n' {
		base.bufPos++
		wasNewline = true
	}

	if wasNewline {
		base.prevLineCharNum = base.charNum
		base.charNum = 1
		base.lineNum++
	}

	return
}

func (base Base) GetLine(n int) (line string) {
	if n <= 0 || n > math.MaxUint32 {
		return
	}

	num := uint32(n)

	isNewLineChar := func(char byte) bool {
		return char == '\r' || char == '\n'
	}

	for base.bufPos < len(base.buf) {
		line = base.TakeUntil(isNewLineChar)

		if base.lineNum >= num {
			break
		}

		base.HandleNewline()
		line = ""
	}

	return
}

func (base *Base) Pos() token.Pos {
	return token.Pos{
		FileID: base.fileID,
		Offset: uint64(base.bufPos),
		Line:   uint32(base.lineNum),
		Char:   uint32(base.charNum),
	}
}

func (base *Base) PrevPos() token.Pos {
	if base.bufPos == 0 {
		return token.Pos{}
	}

	pos := token.Pos{
		FileID: base.fileID,
		Offset: uint64(base.bufPos - 1),
		Line:   uint32(base.lineNum),
		Char:   uint32(base.charNum - 1),
	}

	if pos.Char == 0 {
		pos.Line--
		pos.Char = uint32(base.prevLineCharNum)
	}

	return pos
}
