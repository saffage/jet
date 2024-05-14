package scanner

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/galsondor/go-ascii"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/scanner/base"
	"github.com/saffage/jet/token"
)

func Scan(buffer []byte, fileid config.FileID, flags Flags) ([]token.Token, []error) {
	s := New(buffer, fileid, flags)
	return s.AllTokens(), s.Errors()
}

type Flags int

const (
	NoFlags        Flags = 0
	SkipWhitespace Flags = 1 << (iota - 1)
	SkipIllegal
	SkipComments

	DefaultFlags Flags = NoFlags
)

type Scanner struct {
	*base.Base
	errors []error
	flags  Flags
}

func New(buffer []byte, fileid config.FileID, flags Flags) *Scanner {
	return &Scanner{
		Base:  base.New(buffer, fileid),
		flags: flags,
	}
}

func (s *Scanner) Errors() []error {
	return s.errors
}

func (s *Scanner) AllTokens() []token.Token {
	toks := []token.Token{}

	for {
		tok := s.Next()
		toks = append(toks, tok)

		if tok.Kind == token.EOF {
			break
		}
	}

	return toks
}

func (s *Scanner) Next() token.Token {
	if !s.Match('\000') {
		startPos, tok := s.Pos(), token.Token{Kind: token.Illegal}

		switch {
		case s.Match('#'):
			tok = token.Token{
				Kind: token.Comment,
				Data: s.TakeUntil(isNewLineChar),
			}

			if s.flags&SkipComments != 0 {
				return s.Next()
			}

		case s.Consume('@'):
			tok = token.Token{Kind: token.At}

		case s.Match(' '):
			tok = token.Token{
				Kind: token.Whitespace,
				Data: s.TakeWhile(func(c byte) bool { return c == ' ' }),
			}

		case s.Match('\t'):
			tok = token.Token{
				Kind: token.Tab,
				Data: string(s.Advance()),
			}

		case token.IsIdentifierStartChar(s.Peek()):
			identifier := s.TakeWhile(token.IsIdentifierChar)

			if kind := token.KindFromString(identifier); kind != token.Illegal {
				tok = token.Token{Kind: kind}
			} else {
				tok = token.Token{
					Kind: token.Ident,
					Data: identifier,
				}
			}

		case isNewLineChar(s.Peek()):
			tok = token.Token{
				Kind: token.NewLine,
				Data: s.TakeWhile(isNewLineChar),
			}

		case ascii.IsDigit(s.Peek()):
			tok = s.scanNumber()

		case s.Match('"', '\''):
			tok = s.scanString()

		case s.Consume('.'):
			kind := token.Dot

			if s.Consume('.') {
				if s.Consume('.') {
					kind = token.Ellipsis
				} else if s.Consume('<') {
					kind = token.Dot2Less
				} else {
					kind = token.Dot2
				}
			}

			tok = token.Token{Kind: kind}

		case s.Consume('!', '+', '*', '/', '%'):
			// NOTE This tokens is order dependent
			kind := token.KindFromString(string(s.Prev()))

			if s.Consume('=') {
				kind += 1
			}

			tok = token.Token{Kind: kind}

		case s.Consume('<'):
			kind := token.LtOp

			if s.Consume('<') {
				kind = token.Shl
			} else if s.Consume('=') {
				kind = token.LeOp
			}

			tok = token.Token{Kind: kind}

		case s.Consume('>'):
			kind := token.GtOp

			if s.Consume('>') {
				kind = token.Shr
			} else if s.Consume('=') {
				kind = token.GeOp
			}

			tok = token.Token{Kind: kind}

		case s.Consume('-'):
			kind := token.Minus

			if s.Consume('=') {
				kind = token.MinusEq
			} else if s.Consume('>') {
				kind = token.Arrow
			}

			tok = token.Token{Kind: kind}

		case s.Consume('='):
			kind := token.Eq

			if s.Consume('=') {
				kind = token.EqOp
			} else if s.Consume('>') {
				kind = token.FatArrow
			}

			tok = token.Token{Kind: kind}

		case s.Match('?', '&', '|', '^', ',', ':', ';', '(', ')', '[', ']', '{', '}'):
			kind := token.KindFromString(string(s.Advance()))

			if kind == token.Illegal {
				panic("unreachable")
			}

			tok = token.Token{Kind: kind}

		default:
			s.error("illegal character", s.Pos())

			tok = token.Token{
				Kind: token.Illegal,
				Data: string(s.Advance()),
			}
		}

		if tok.Start.Line == 0 {
			tok.Start = startPos
			tok.End = s.PrevPos()
		}

		if s.flags&SkipWhitespace != 0 &&
			(tok.Kind == token.Whitespace || tok.Kind == token.Tab) {
			return s.Next()
		} else if s.flags&SkipIllegal != 0 && tok.Kind == token.Illegal {
			return s.Next()
		}

		return tok
	}

	return token.Token{
		Kind:  token.EOF,
		Start: s.Pos(),
		End:   s.Pos(),
	}
}

func (s *Scanner) scanString() token.Token {
	quotePos, quote := s.Pos(), s.Advance()
	data := s.Take(func() (data []byte, stop bool) {
		if s.Peek() == '\000' || s.Peek() == quote || isNewLineChar(s.Peek()) {
			return nil, true
		} else if s.Consume('\\') {
			switch s.Advance() {
			case 'n':
				data = []byte{'\n'}

			case 'r':
				data = []byte{'\r'}

			case 't':
				data = []byte{'\t'}

			case '\\':
				data = []byte{'\\'}

			case '\'':
				data = []byte{'\''}

			case '"':
				data = []byte{'"'}

			case 'x':
				if bytes, ok := s.parseBytes(2); !ok {
					data = append([]byte{'\\', 'x'}, bytes...)
				} else {
					data = bytes
				}

			case 'u':
				if bytes, ok := s.parseBytes(4); !ok {
					data = append([]byte{'\\', 'u'}, bytes...)
				} else {
					data = bytes
				}

			case 'U':
				if bytes, ok := s.parseBytes(8); !ok {
					data = append([]byte{'\\', 'U'}, bytes...)
				} else {
					data = bytes
				}

			default:
				s.errorUnexpected(fmt.Sprintf("character escape `\\%c`", s.Prev()), s.PrevPos())
				data = []byte{'\\', s.Prev()}
			}
		} else {
			data = []byte{s.Advance()}
		}

		return
	})

	if !s.Consume(quote) {
		s.error("unterminated string literal", quotePos)
		return token.Token{
			Kind:  token.Illegal,
			Data:  data,
			Start: quotePos,
			End:   s.PrevPos(),
		}
	}

	return token.Token{
		Kind:  token.String,
		Data:  data,
		Start: quotePos,
		End:   s.PrevPos(),
	}
}

func (s *Scanner) parseBytes(n int) ([]byte, bool) {
	i, startPos := 0, s.Pos()
	bytes, realBytes := make([]byte, n), make([]byte, 0, n*2)

	for ; i < n; i++ {
		if s.Consume('_') {
			realBytes = append(realBytes, '_')

			if !ascii.IsHexDigit(s.Peek()) {
				s.error("invalid byte", s.Pos())
				return realBytes, false
			}
		}

		if !ascii.IsHexDigit(s.Peek()) {
			break
		}

		char := s.Advance()

		bytes[i] = char
		realBytes = append(realBytes, char)
	}

	if i == 0 {
		s.error("invalid byte", startPos)
		return realBytes, false
	} else if i < n {
		s.error(fmt.Sprintf("invalid byte (expected %d bytes)", n), s.Pos())
		return realBytes, false
	}

	result := make([]byte, n/2)

	if _, err := hex.Decode(result, bytes); err != nil {
		s.error(err.Error(), s.Pos())
		return nil, false
	}

	return result, true
}

func (s *Scanner) scanNumber() token.Token {
	numPart, fracPart, expPart := "", "", ""

	if s.Consume('0') {
		switch {
		case s.Consume('x'):
			num := s.parseHexNumber()
			num.Data = "0x" + num.Data
			return num

		case s.Consume('b'):
			num := s.parseBinNumber()
			num.Data = "0b" + num.Data
			return num

		case s.Consume('o'):
			num := s.parseOctNumber()
			num.Data = "0o" + num.Data
			return num

		case s.Match('X', 'B', 'O'):
			s.error("uppercase letters is not allowed, use lowercase instead", s.Pos())
			return token.Token{
				Kind: token.Illegal,
				Data: string(s.Peek()),
			}

		case ascii.IsDigit(s.Peek()):
			s.error("`0` as the first character of a number literal is not allowed", s.Pos())
			return token.Token{
				Kind: token.Illegal,
				Data: string(s.Prev()),
			}

		default:
			numPart = "0"
		}
	} else {
		if num := s.parseDecNumber(); num.Kind != token.Illegal {
			numPart = num.Data
		} else {
			return num
		}
	}

	if s.Match('.') && ascii.IsDigit(s.LookAhead(1)) {
		s.Advance()
		fracPart = "."

		if num := s.parseNumber(ascii.IsDigit, "number after the point"); num.Kind != token.Illegal {
			fracPart += num.Data
		} else {
			return num
		}
	}

	if s.Consume('e', 'E') {
		expPart = string(s.Prev())

		if s.Consume('+', '-') {
			expPart += string(s.Prev())
		}

		if num := s.parseDecNumber(); num.Kind != token.Illegal {
			expPart += num.Data
		} else {
			return num
		}
	}

	if token.IsIdentifierStartChar(s.Peek()) {
		s.errorUnexpected("character", s.Pos(), "numeric literals have no suffixes")
		return token.Token{
			Kind: token.Illegal,
			Data: string(s.Peek()),
		}
	}

	tok := token.Token{
		Kind: token.Int,
		Data: numPart + fracPart + expPart,
	}

	if len(fracPart) > 0 || len(expPart) > 0 {
		tok.Kind = token.Float
	}

	return tok
}

// On success returns token of kind [token.Int], otherwise returns [token.Illegal] is cases:
//
//   - if first character is not a <number>
//   - if character after '_' is not a <number>
//
// Pattern is: `<number> ('_' <number>)*`
func (s *Scanner) parseNumber(predicate func(byte) bool, expected string) token.Token {
	if !predicate(s.Peek()) {
		s.errorExpected(expected, s.Pos())
		return token.Token{
			Kind: token.Illegal,
			Data: string(s.Peek()),
		}
	}

	num := strings.Builder{}
	num.WriteByte(s.Advance())

	for {
		if s.Consume('_') {
			num.WriteByte('_')

			if !predicate(s.Peek()) {
				num.WriteByte(s.Peek())
				s.errorExpected(expected+" after `_`", s.Pos())
				return token.Token{
					Kind: token.Illegal,
					Data: num.String(),
				}
			}
		} else if !predicate(s.Peek()) {
			break
		}

		num.WriteByte(s.Advance())
	}

	return token.Token{
		Kind: token.Int,
		Data: num.String(),
	}
}

func (s *Scanner) parseBinNumber() token.Token {
	return s.parseNumber(isBinNumChar, "binary number")
}

func (s *Scanner) parseOctNumber() token.Token {
	return s.parseNumber(isOctNumChar, "octal number")
}

func (s *Scanner) parseDecNumber() token.Token {
	return s.parseNumber(ascii.IsDigit, "number")
}

func (s *Scanner) parseHexNumber() token.Token {
	return s.parseNumber(ascii.IsHexDigit, "hexadecimal number")
}

func isBinNumChar(char byte) bool {
	return char == '0' || char == '1'
}

func isOctNumChar(char byte) bool {
	return '0' <= char && char <= '7'
}

func isNewLineChar(char byte) bool {
	return char == '\r' || char == '\n'
}
