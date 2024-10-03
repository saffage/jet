package scanner

import (
	"errors"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser/token"
)

type Flags int

const (
	SkipIllegal Flags = 1 << iota
	SkipComments

	NoFlags      Flags = 0
	DefaultFlags Flags = SkipComments
)

func Scan(buffer []byte, fileid config.FileID, flags Flags) ([]token.Token, error) {
	s := New(buffer, fileid, flags)
	return s.AllTokens(), errors.Join(s.errors...)
}

func MustScan(buffer []byte, fileid config.FileID, flags Flags) []token.Token {
	tokens, err := Scan(buffer, fileid, flags)
	if err != nil {
		panic(err)
	}
	return tokens
}
