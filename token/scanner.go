package token

import (
	"encoding/hex"
	"errors"
	"fmt"
	_ "go/token"
	"strings"
	"unicode"

	"github.com/saffage/jet/text"
)

type ScannerFlags int

const (
	SkipIllegal ScannerFlags = 1 << iota
	SkipComments

	NoFlags      ScannerFlags = 0
	DefaultFlags ScannerFlags = NoFlags
)

var (
	ErrIllegalCharacter        = errors.New("illegal character")
	ErrIllegalNumericBase      = errors.New("invalid numeric base")
	ErrInvalidByte             = errors.New("invalid byte")
	ErrInvalidEscape           = errors.New("invalid character escape")
	ErrUnterminatedStringLit   = errors.New("unterminated string literal")
	ErrFirstDigitIsZero        = errors.New("'0' as the first digit of a number literal is not allowed")
	ErrExpectedIdentForSuffix  = errors.New("expected identifier for numeric suffix")
	ErrExpectedDigitAfterPoint = errors.New("expected digit after the decimal point")
	ErrExpectedDecNumber       = errors.New("expected decimal number")
	ErrExpectedHexNumber       = errors.New("expected hexadecimal number")
	ErrExpectedBinNumber       = errors.New("expected binary number")
	ErrExpectedOctNumber       = errors.New("expected octal number")
	ErrExpectedByte            = errors.New("expected byte")
)

type Scanner struct {
	text.Scanner
	errors []error
	flags  ScannerFlags
}

func New(input []byte, id text.FileID, flags ScannerFlags) *Scanner {
	return &Scanner{
		Scanner: *text.NewScanner(input, id),
		flags:   flags,
	}
}

func Scan(input []byte, id text.FileID, flags ScannerFlags) ([]Token, error) {
	s := New(input, id, flags)
	return s.AllTokens(), errors.Join(s.errors...)
}

func (s *Scanner) AllTokens() (tokens []Token) {
	for {
		tok := s.NextToken()
		tokens = append(tokens, tok)

		if tok.Kind == EOF {
			break
		}
	}
	return
}

func (s *Scanner) NextToken() Token {
	if !s.Match('\000') {
		startPos, tok := s.Pos(), Token{Kind: Illegal}

		s.SkipWhile(isSpace)

		switch {
		case s.Match('#'):
			tok = Token{
				Kind: Comment,
				Data: s.TakeUntil(isNewLine),
			}

			if s.flags&SkipComments != 0 {
				return s.NextToken()
			}

		case s.Consumed('@'):
			tok = Token{Kind: At}

		case s.Consumed('$'):
			tok = Token{Kind: Dollar}

		case s.Match('\n', '\r'):
			s.TakeWhile(isNewLine)

			tok = Token{
				Kind: Semicolon,
				Data: "\n",
			}

		case isDigit(s.Peek()):
			tok = s.scanNumber()

		case IsIdentifierStartChar(s.Peek()):
			identifier := s.TakeWhile(IsIdentifierChar)

			if kind := KindFrom(identifier); kind != Illegal {
				tok = Token{Kind: kind}
			} else {
				tok = Token{
					Kind: Ident,
					Data: identifier,
				}

				if s.Match('"', '\'') {
					strTok := s.scanString()
					strTok.Span.From = tok.Span.From
					strTok.Data = tok.Data + strTok.Data
					tok = strTok
				}
			}

		case s.Match('"', '\''):
			tok = s.scanString()

		case s.Consumed('.'):
			kind := Dot

			if s.Consumed('.') {
				if s.Consumed('.') {
					kind = Ellipsis
				} else if s.Consumed('<') {
					kind = Dot2Less
				} else {
					kind = Dot2
				}
			}

			tok = Token{Kind: kind}

		case s.Match('!', '+', '*', '/', '%', '&', '|', '^'):
			// NOTE This tokens is order dependent
			kind := KindFrom(s.Advance())

			if s.Consumed('=') {
				kind += 1
			}

			tok = Token{Kind: kind}

		case s.Consumed('<'):
			kind := LtOp

			if s.Consumed('<') {
				kind = Shl
			} else if s.Consumed('=') {
				kind = LeOp
			}

			tok = Token{Kind: kind}

		case s.Consumed('>'):
			kind := GtOp

			if s.Consumed('>') {
				kind = Shr
			} else if s.Consumed('=') {
				kind = GeOp
			}

			tok = Token{Kind: kind}

		case s.Consumed('-'):
			kind := Minus

			if s.Consumed('=') {
				kind = MinusEq
			} else if s.Consumed('>') {
				kind = Arrow
			}

			tok = Token{Kind: kind}

		case s.Consumed('='):
			kind := Eq

			if s.Consumed('=') {
				kind = EqOp
			} else if s.Consumed('>') {
				kind = FatArrow
			}

			tok = Token{Kind: kind}

		case s.Match(',', ':', ';', '(', ')', '[', ']', '{', '}'):
			kind := KindFrom(s.Advance())

			if kind == Illegal {
				panic("unreachable")
			}

			tok = Token{Kind: kind}

		default:
			s.error(ErrIllegalCharacter, s.Pos())

			tok = Token{
				Kind: Illegal,
				Data: string(s.Advance()),
			}
		}

		if !tok.Span.From.IsValid() {
			tok.Span.From = startPos
		}

		if !tok.Span.To.IsValid() {
			pos := s.Pos()
			tok.Span.To = text.PosFrom(pos.ID(), pos.Offset()-1)
		}

		if s.flags&SkipIllegal != 0 && tok.Kind == Illegal {
			return s.NextToken()
		}

		return tok
	}

	return Token{
		Kind: EOF,
		Span: text.Span{From: s.Pos(), To: s.Pos()},
	}
}

func (s *Scanner) scanString() Token {
	openingQuotePos := s.Pos()
	quote := s.ExpectChar('"', '\'')
	data := strings.Builder{}

	for {
		switch char := s.Peek(); char {
		case quote:
			closingQuotePos := s.Pos()
			s.NextToken()
			data.WriteRune(quote)

			return Token{
				Kind: String,
				Data: data.String(),
				Span: text.Span{From: openingQuotePos, To: closingQuotePos},
			}

		case '\000', '\n', '\r':
			s.error(ErrUnterminatedStringLit, openingQuotePos)

			return Token{
				Kind: Illegal,
				Data: data.String(),
				Span: text.Span{From: openingQuotePos, To: s.Pos()},
			}

		case '\\':
			backslashPos := s.Pos()

			switch char = s.Next(); char {
			case 'n':
				data.WriteByte('\n')

			case 'r':
				data.WriteByte('\r')

			case 't':
				data.WriteByte('\t')

			case '\\':
				data.WriteByte('\\')

			case '\'':
				data.WriteByte('\'')

			case '"':
				data.WriteByte('"')

			case 'x':
				if !s.parseBytes(&data, 2) {
					// TODO error
				}

			case 'u':
				if !s.parseBytes(&data, 4) {
					// TODO error
				}

			case 'U':
				if !s.parseBytes(&data, 8) {
					// TODO error
				}

			default:
				s.error(ErrInvalidEscape, backslashPos)

				// NOTE not sure if invalid escape needs to be present in token
				data.WriteByte('\\')
				data.WriteRune(char)
			}

		case '$':
			// TODO interpolated string

		default:
			data.WriteRune(s.Advance())
		}
	}
}

func (s *Scanner) scanNumber() Token {
	buf := strings.Builder{}
	tok := Token{Kind: Int}
	begin := s.Pos()

	if s.Consumed('0') {
		if char, consumed := s.Consume('x', 'X', 'b', 'B', 'o', 'O'); consumed {
			if unicode.IsUpper(char) {
				// TODO warning?
				s.error(
					ErrIllegalNumericBase,
					s.Pos(),
					"uppercase letters in numeric base prefix is not allowed, use lowercase letter instead",
				)
			}

			char = unicode.ToLower(char)
			buf.WriteByte('0')
			buf.WriteRune(char)

			switch char {
			case 'x':
				s.parseHexNumber(&buf)

			case 'b':
				s.parseBinNumber(&buf)

			case 'o':
				s.parseOctNumber(&buf)

			default:
				panic("unreachable")
			}

			tok.Data = buf.String()
			return tok
		}

		s.parseNumber(nil, isZero, nil)

		if s.SkipWhile(isZero) > 0 {
			// TODO warning?
			s.error(ErrFirstDigitIsZero, begin)
		}

		s.ConsumeFunc(isDigit)

		switch {
		case isDigit(s.Peek()):
			// TODO parse it like regular number
			s.error(ErrFirstDigitIsZero, begin)
			return Token{
				Kind: Illegal,
				Data: "0",
			}

		default:
			buf.WriteByte('0')
		}
	} else if !s.parseDecNumber(&buf) {
		return Token{Kind: Illegal}
	}

	if s.Consumed('.') {
		buf.WriteByte('.')
		tok.Kind = Float

		if !s.parseNumber(&buf, isDigit, ErrExpectedDigitAfterPoint) {
			return Token{Kind: Illegal}
		}
	}

	if s.Match('e', 'E') {
		buf.WriteRune(s.Advance())
		tok.Kind = Float

		if s.Match('+', '-') {
			buf.WriteRune(s.Advance())
		}

		if !s.parseDecNumber(&buf) {
			return Token{Kind: Illegal}
		}
	}

	// Numeric Suffix
	//
	// if s.Consumed('\'') {
	// 	buf.WriteByte('\'')
	//
	// 	if !IsIdentifierStartChar(s.Peek()) {
	// 		s.error(ErrExpectedIdentForSuffix, s.Pos())
	// 		tok.Kind = Illegal
	// 		return tok
	// 	} else {
	// 		buf.WriteString(s.TakeWhile(IsIdentifierChar))
	// 	}
	// }

	tok.Data = buf.String()
	return tok
}

// Returns false when:
//
//   - first character is not a byte or '_'
//   - character after '_' is not a byte
//   - number of bytes not equals to n
//
// Pattern is:
//
//	{'_' number}
func (s *Scanner) parseBytes(buf *strings.Builder, n int) bool {
	if n < 0 || n > 8 || n%2 != 0 {
		panic("invalid argument")
	}

	bytes := [8]byte{}
	startPos := s.Pos()
	wasUnderscore := false

	for i := 0; i < n; i++ {
		wasUnderscore = s.Consumed('_')
		char, consumed := s.ConsumeFunc(isHexDigit)

		if !consumed {
			if wasUnderscore {
				s.error(ErrInvalidByte, s.Pos(), "unexpected character after underscore")
				return false
			}

			if i == 0 {
				s.error(ErrExpectedByte, startPos)
				return false
			}

			if i < n {
				s.error(ErrInvalidByte, s.Pos(), "expected ", n, " bytes")
				return false
			}

			break
		}

		bytes[i] = byte(char)
	}

	result := [4]byte{}
	_, err := hex.Decode(result[:], bytes[:n])

	if err != nil {
		// Unreachable, because we can't parse bytes that is not
		// a hexadecimal number.
		panic(err)
	}

	if buf != nil {
		buf.Write(result[:n/2])
	}

	return true
}

// Returns false when:
//
//   - first character is not a number
//   - character after '_' is not a number
//
// Pattern is:
//
//	number {'_' number}
func (s *Scanner) parseNumber(buf *strings.Builder, predicate func(rune) bool, err error) bool {
	written := 0
	wasUnderscore := false

	for {
		char, consumed := s.ConsumeFunc(predicate)

		if !consumed {
			if wasUnderscore {
				s.error(err, s.Pos(), "unexpected character after underscore")
				return false
			}

			if written == 0 {
				s.error(err, s.Pos())
				return false
			}

			return true
		}

		if buf != nil {
			buf.WriteRune(char)
		}

		written++
		wasUnderscore = s.Consumed('_')
	}
}

func (s *Scanner) parseBinNumber(buf *strings.Builder) bool {
	return s.parseNumber(buf, isBinDigit, ErrExpectedBinNumber)
}

func (s *Scanner) parseOctNumber(buf *strings.Builder) bool {
	return s.parseNumber(buf, isOctDigit, ErrExpectedOctNumber)
}

func (s *Scanner) parseDecNumber(buf *strings.Builder) bool {
	return s.parseNumber(buf, isDigit, ErrExpectedDecNumber)
}

func (s *Scanner) parseHexNumber(buf *strings.Builder) bool {
	return s.parseNumber(buf, isHexDigit, ErrExpectedHexNumber)
}

// Emits an error. Error end is a current scanner position.
func (s *Scanner) error(err error, start text.Pos, message ...any) {
	if err != nil {
		s.errors = append(s.errors, &Error{
			Message:   fmt.Sprint(message...),
			Selection: text.Span{From: start, To: s.Pos()},
			err:       err,
		})
	}
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func isHexDigit(c rune) bool {
	return '0' <= c && c <= '9' ||
		'a' <= c && c <= 'f' ||
		'A' <= c && c <= 'F'
}

func isBinDigit(c rune) bool {
	return c == '0' || c == '1'
}

func isOctDigit(c rune) bool {
	return '0' <= c && c <= '7'
}

func isNewLine(c rune) bool {
	return c == '\r' || c == '\n'
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t'
}

func isZero(c rune) bool {
	return c == '0'
}
