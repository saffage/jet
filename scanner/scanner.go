package scanner

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/galsondor/go-ascii"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/scanner/base"
	"github.com/saffage/jet/token"
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

func (s *Scanner) AllTokens() (toks []token.Token) {
	for {
		tok := s.Next()
		toks = append(toks, tok)

		if tok.Kind == token.EOF {
			break
		}
	}
	return
}

func (s *Scanner) Next() token.Token {
	if !s.Match('\000') {
		var startPos = s.Pos()
		var tok token.Token

		switch {
		case s.ConsumeAny(' ', '\t', '\n', '\r'):
			// s.TakeWhile(func(b byte) bool {
			// 	return bytes.IndexByte([]byte{' ', '\t', '\n', '\r'}, b) >= 0
			// })
			return s.Next()

		case s.Match('#'):
			tok = token.Token{
				Kind: token.Comment,
				Data: s.TakeUntil(isNewLine),
			}

			if s.flags&SkipComments != 0 {
				return s.Next()
			}

		case ascii.IsAlnum(s.Peek()), s.Match('_'):
			ident := s.TakeWhile(token.IsIdentChar)

			if kind := token.KindFromString(ident); kind != token.Illegal {
				tok = token.Token{Kind: kind}
			} else if ident[0] == '_' {
				tok = token.Token{Kind: token.Underscore, Data: ident}
			} else if strings.ContainsFunc(ident, unicode.IsUpper) {
				tok = token.Token{Kind: token.Type, Data: ident}
			} else {
				tok = token.Token{Kind: token.Ident, Data: ident}
			}

		case ascii.IsDigit(s.Peek()):
			tok = s.scanNumber()

		case s.Match('"', '\''):
			tok = s.scanString()

		case s.Consume('.'):
			tok = token.Token{Kind: token.Dot}

			if s.Consume('.') {
				tok.Kind = token.Dot2
			}

		case s.ConsumeAny('!', '+', '*', '/', '%', '^'):
			// NOTE This tokens is order dependent
			tok = token.Token{Kind: token.KindFromByte(s.Prev())}

			if s.Consume('=') {
				tok.Kind++
			}

		case s.Consume('|'):
			tok = token.Token{Kind: token.Pipe}

			if s.Consume('=') {
				tok.Kind = token.PipeEq
			} else if s.Consume('|') {
				tok.Kind = token.Or

				if s.Consume('=') {
					tok.Kind = token.OrEq
				}
			}

		case s.Consume('&'):
			tok = token.Token{Kind: token.Amp}

			if s.Consume('=') {
				tok.Kind = token.AmpEq
			} else if s.Consume('|') {
				tok.Kind = token.And

				if s.Consume('=') {
					tok.Kind = token.AndEq
				}
			}

		case s.Consume('<'):
			tok = token.Token{Kind: token.LtOp}

			if s.Consume('=') {
				tok.Kind = token.LeOp
			} else if s.Consume('<') {
				tok.Kind = token.Shl

				if s.Consume('=') {
					tok.Kind = token.ShlEq
				}
			}

		case s.Consume('>'):
			tok = token.Token{Kind: token.GtOp}

			if s.Consume('=') {
				tok.Kind = token.GeOp
			} else if s.Consume('>') {
				tok.Kind = token.Shr

				if s.Consume('=') {
					tok.Kind = token.ShrEq
				}
			}

		case s.Consume('-'):
			tok = token.Token{Kind: token.Minus}

			if s.Consume('=') {
				tok.Kind = token.MinusEq
			} else if s.Consume('>') {
				tok.Kind = token.Arrow
			}

		case s.Consume('='):
			tok = token.Token{Kind: token.Eq}

			if s.Consume('=') {
				tok.Kind = token.EqOp
			} else if s.Consume('>') {
				tok.Kind = token.FatArrow
			}

		case s.ConsumeAny('(', ')', '{', '}', '[', ']', ':', ','):
			tok = token.Token{Kind: token.KindFromByte(s.Prev())}

			if tok.Kind == token.Illegal {
				panic("unreachable")
			}

		default:
			s.error(ErrIllegalCharacter, s.Pos())

			tok = token.Token{
				Kind: token.Illegal,
				Data: string(s.Advance()),
			}
		}

		if !tok.Start.IsValid() {
			tok.Start = startPos
		}

		if !tok.End.IsValid() {
			tok.End = s.PrevPos()
		}

		if s.flags&SkipIllegal != 0 && tok.Kind == token.Illegal {
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
	quotePos := s.Pos()
	quote := s.Advance()
	data := string(quote) + s.Take(func() (data []byte, stop bool) {
		if s.Peek() == '\000' || s.Peek() == quote || isNewLine(s.Peek()) {
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
				s.error(
					ErrInvalidEscape,
					s.PrevPos(),
					fmt.Sprintf("'\\%c'", s.Prev()),
				)
				data = []byte{'\\', s.Prev()}
			}
		} else {
			data = []byte{s.Advance()}
		}

		return
	})

	if !s.Consume(quote) {
		s.error(ErrUnterminatedStringLit, quotePos)
		return token.Token{
			Kind:  token.Illegal,
			Data:  data,
			Start: quotePos,
			End:   s.PrevPos(),
		}
	}

	return token.Token{
		Kind:  token.String,
		Data:  data + string(quote),
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
				s.error(ErrInvalidByte, s.Pos())
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
		s.error(ErrInvalidByte, startPos)
		return realBytes, false
	} else if i < n {
		s.error(ErrInvalidByte, s.Pos(), fmt.Sprintf("expected %d bytes", n))
		return realBytes, false
	}

	result := make([]byte, n/2)

	if _, err := hex.Decode(result, bytes); err != nil {
		s.error(err, s.Pos())
		return nil, false
	}

	return result, true
}

func (s *Scanner) scanNumber() token.Token {
	buf := strings.Builder{}
	tok := token.Token{Kind: token.Int, Data: ""}

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
			s.error(ErrIllegalNumericBase, s.Pos())
			return token.Token{
				Kind: token.Illegal,
				Data: string(s.Peek()),
			}

		case ascii.IsDigit(s.Peek()):
			s.error(ErrFirstDigitIsZero, s.Pos())
			return token.Token{
				Kind: token.Illegal,
				Data: string(s.Prev()),
			}

		default:
			buf.WriteByte('0')
		}
	} else {
		if num := s.parseDecNumber(); num.Kind != token.Illegal {
			buf.WriteString(num.Data)
		} else {
			return num
		}
	}

	if s.Match('.') && ascii.IsDigit(s.LookAhead(1)) {
		s.Advance()
		buf.WriteByte('.')
		tok.Kind = token.Float

		num := s.parseNumber(ascii.IsDigit, ErrExpectedDigitAfterPoint)
		if num.Kind != token.Illegal {
			buf.WriteString(num.Data)
		} else {
			return num
		}
	}

	if s.ConsumeAny('e', 'E') {
		buf.WriteByte(s.Prev())
		tok.Kind = token.Float

		if s.ConsumeAny('+', '-') {
			buf.WriteByte(s.Prev())
		}

		if num := s.parseDecNumber(); num.Kind != token.Illegal {
			buf.WriteString(num.Data)
		} else {
			return num
		}
	}

	if s.Consume('\'') {
		buf.WriteByte('\'')

		if !token.IsIdentStartChar(s.Peek()) {
			s.error(ErrExpectedIdentForSuffix, s.Pos())
			tok.Kind = token.Illegal
			return tok
		} else {
			buf.WriteString(s.TakeWhile(token.IsIdentChar))
		}
	}

	tok.Data = buf.String()
	return tok
}

// On success returns token of kind [token.Int], otherwise returns [token.Illegal] is cases:
//
//   - if first character is not a <number>
//   - if character after '_' is not a <number>
//
// Pattern is: `<number> ('_' <number>)*`
func (s *Scanner) parseNumber(predicate func(byte) bool, err error) token.Token {
	if !predicate(s.Peek()) {
		s.error(err, s.Pos())
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
				s.error(err, s.Pos())
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
	return s.parseNumber(isBinDigit, ErrExpectedBinNumber)
}

func (s *Scanner) parseOctNumber() token.Token {
	return s.parseNumber(isOctDigit, ErrExpectedOctNumber)
}

func (s *Scanner) parseDecNumber() token.Token {
	return s.parseNumber(ascii.IsDigit, ErrExpectedDecNumber)
}

func (s *Scanner) parseHexNumber() token.Token {
	return s.parseNumber(ascii.IsHexDigit, ErrExpectedHexNumber)
}

func isBinDigit(c byte) bool {
	return c == '0' || c == '1'
}

func isOctDigit(c byte) bool {
	return '0' <= c && c <= '7'
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}
