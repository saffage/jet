package text

import (
	"bytes"
	"io"
	"iter"
	"slices"
	"strings"
)

const eof = '\000'

type UnexpectedCharError struct {
	Expected string
	Found    rune
	Range    Span
}

type Scanner struct {
	reader bytes.Reader
	fileID FileID
	peeked rune
	index  int
	size   int
}

func NewScanner(input []byte, id FileID) *Scanner {
	s := &Scanner{
		reader: *bytes.NewReader(input),
		fileID: id,
	}
	s.Peek()
	return s
}

func (s *Scanner) Next() (peeked rune) {
	char, size, err := s.reader.ReadRune()

	if err != nil {
		if err == io.EOF {
			if s.peeked != eof {
				s.peeked = eof
				s.index += 1
				s.size = 0
			}
			return eof
		}
		// Unreachable because [bytes.Reader.ReadRune] can't return other error.
		panic(err)
	}

	s.peeked = char
	s.index += s.size
	s.size = size
	return char
}

// TODO rename, `Advance` is not clear enough
func (s *Scanner) Advance() (previous rune) {
	previous = s.peeked
	s.Next()
	return previous
}

func (s *Scanner) Peek() (peeked rune) {
	if s.peeked == eof {
		// Likely first read. Or last.
		s.Next()
	}

	return s.peeked
}

func (s *Scanner) ExpectFunc(predicate func(peeked rune) (expected string, ok bool)) rune {
	char := s.Peek()

	if expected, ok := predicate(char); !ok {
		panic(UnexpectedCharError{
			Range:    Span{From: s.Pos()},
			Found:    char,
			Expected: expected,
		})
	}

	s.Next()
	return char
}

func (s *Scanner) ExpectChar(chars ...rune) rune {
	char, ok := s.Consume(chars...)

	if !ok {
		expected := []string{}

		for _, char := range chars {
			expected = append(expected, string(char))
		}

		panic(UnexpectedCharError{
			Range:    Span{From: s.Pos()},
			Found:    char,
			Expected: strings.Join(expected, ", "),
		})
	}

	return char
}

func (s *Scanner) Expect(what string, chars ...rune) rune {
	char, ok := s.Consume(chars...)

	if !ok {
		panic(UnexpectedCharError{
			Range:    Span{From: s.Pos()},
			Found:    char,
			Expected: what,
		})
	}

	return char
}

func (s *Scanner) Consume(chars ...rune) (char rune, consumed bool) {
	char = s.Peek()

	if len(chars) > 0 && slices.Contains(chars, char) {
		s.Next()
		consumed = true
	}

	return
}

func (s *Scanner) ConsumeFunc(predicate func(peeked rune) bool) (char rune, consumed bool) {
	char = s.Peek()

	if predicate(char) {
		s.Next()
		consumed = true
	}

	return
}

func (s *Scanner) Consumed(chars ...rune) bool {
	_, consumed := s.Consume(chars...)
	return consumed
}

//
//
//

// Returns true if the current character matches `char`.
func (s *Scanner) Match(chars ...rune) bool {
	return slices.Contains(chars, s.Peek())
}

func (s *Scanner) Chars() iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for {
			if char := s.Peek(); !yield(char) || char == eof {
				return
			}
			s.Advance()
		}
	}
}

// Takes all characters while predicate returns true.
func (s *Scanner) TakeWhile(predicate func(rune) bool) string {
	buf := strings.Builder{}
	for char := range s.Chars() {
		if !predicate(char) {
			break
		}
		buf.WriteRune(char)
	}
	return buf.String()
}

// Takes all characters while predicate returns false.
func (s *Scanner) TakeUntil(predicate func(rune) bool) string {
	buf := strings.Builder{}
	for char := range s.Chars() {
		if predicate(char) {
			break
		}
		buf.WriteRune(char)
	}
	return buf.String()
}

// Skips all characters while predicate returns true.
func (s *Scanner) SkipWhile(predicate func(rune) bool) (skipped int) {
	for char := range s.Chars() {
		if !predicate(char) {
			break
		}
		skipped++
	}
	return
}

// Skips all characters while predicate returns false.
func (s *Scanner) SkipUntil(predicate func(rune) bool) (skipped int) {
	for char := range s.Chars() {
		if predicate(char) {
			break
		}
		skipped++
	}
	return
}

func wrapPredicate(s *Scanner, predicate func(rune) bool, truth bool) func() (string, bool) {
	return func() (data string, stop bool) {
		if predicate(s.Peek()) == truth {
			data = string(s.peeked)
			s.Next()
		} else {
			stop = true
		}
		return
	}
}

func (s *Scanner) Pos() Pos {
	if s.fileID == 0 {
		panic("scanner: called Pos() with invalid file ID")
	}
	return PosFrom(s.fileID, s.index)
}
