package scanner

import (
	"errors"

	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

type Flags int

const (
	SkipWhitespace Flags = 1 << iota
	SkipIllegal
	SkipComments

	NoFlags      Flags = 0
	DefaultFlags Flags = NoFlags
)

func Scan(buffer []byte, id text.FileID, flags Flags) ([]token.Token, error) {
	s := New(buffer, id, flags)
	return s.AllTokens(), errors.Join(s.errors...)
}
