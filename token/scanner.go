package token

import (
	"encoding/hex"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
	"unicode"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

type ScannerFlags int

const (
	SkipIllegal ScannerFlags = 1 << iota
	SkipComments
	SkipSemicolons
	SkipNewlines

	NoFlags      ScannerFlags = 0
	DefaultFlags ScannerFlags = SkipComments
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
	errors []error
	text.Scanner
	flags ScannerFlags

	emitNewLine bool
}

// TODO remove it
type Token struct {
	Data string
	Kind Kind
	Span text.Span
}

func NewScanner(content []byte, id text.FileID, flags ScannerFlags) *Scanner {
	return &Scanner{
		Scanner: *text.NewScanner(content, id),
		flags:   flags,
	}
}

func NewScannerFromFile(file *text.File, flags ScannerFlags) *Scanner {
	return &Scanner{
		Scanner: *text.NewScannerFromFile(file),
		flags:   flags,
	}
}

func Scan(input []byte, id text.FileID, flags ScannerFlags) ([]Token, error) {
	s := NewScanner(input, id, flags)
	return slices.Collect(s.Tokens()), report.Join(s.errors...)
}

func (s *Scanner) Tokens() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for {
			if tok := s.NextToken(); !yield(tok) || tok.Kind == EOF {
				return
			}
		}
	}
}

func (s *Scanner) NextToken() Token {
	if !s.Match('\000') {
		s.SkipWhile(isSpace)

		pos := s.Pos()
		kind := Illegal
		data := ""
		span := text.Span{From: pos, To: pos}

		if !span.IsValid() {
			panic("unreachable")
		}

		switch {
		case s.Match('#'):
			kind = Comment
			data = s.TakeUntil(isNewLine)

			if s.flags&SkipComments != 0 {
				return s.NextToken()
			}

		case s.Consumed('@'):
			kind = At

		case s.Consumed('$'):
			kind = Dollar

		case s.Match('\n', '\r'):
			s.TakeWhile(isNewLine)

			if s.flags&SkipNewlines != 0 || !s.emitNewLine {
				return s.NextToken()
			}

			kind = Newline

		case s.Consumed(';'):
			if s.flags&SkipSemicolons != 0 {
				return s.NextToken()
			}

			kind = Semicolon

		case isDigit(s.Peek()):
			var ok bool
			kind, data, span, ok = s.scanNumber()

			if !ok {
				kind = Illegal
			}

		case IsIdentifierStartChar(s.Peek()):
			identifier := s.TakeWhile(IsIdentifierChar)
			kind = FromRepresentation(identifier)

			if kind == Illegal {
				data = identifier
				first := []rune(identifier)[0]

				switch {
				case first == '_':
					kind = IdentPlaceholder

				case unicode.IsUpper(first):
					kind = UppercaseIdent

				case unicode.IsLower(first):
					kind = LowercaseIdent

					if s.Match('"', '\'') {
						strData, strSpan, ok := s.scanString()
						if ok {
							data += strData
							span.To = strSpan.To
							kind = String
						} else {
							kind = Illegal
						}
					}
				}
			}

		case s.Match('"', '\''):
			var ok bool
			data, span, ok = s.scanString()
			kind = String

			if !ok {
				kind = Illegal
			}

		case s.Consumed('.'):
			kind = Dot

			if s.Consumed('.') {
				if s.Consumed('.') {
					kind = Ellipsis
				} else {
					kind = Dot2
				}
			}

		case s.Match('!', '+', '*', '/', '%', '&', '|', '^'):
			kind = FromRepresentation(s.Advance())

			if kind == Illegal {
				panic("unreachable")
			}

			// NOTE This tokens is order dependent
			if s.Consumed('=') {
				kind += 1
			}

		case s.Consumed('<'):
			kind = LtOp

			if s.Consumed('<') {
				kind = Shl
			} else if s.Consumed('=') {
				kind = LeOp
			}

		case s.Consumed('>'):
			kind = GtOp

			if s.Consumed('>') {
				kind = Shr
			} else if s.Consumed('=') {
				kind = GeOp
			}

		case s.Consumed('-'):
			kind = Minus

			if s.Consumed('=') {
				kind = MinusEq
			} else if s.Consumed('>') {
				kind = Arrow
			}

		case s.Consumed('='):
			kind = Eq

			if s.Consumed('=') {
				kind = EqOp
			} else if s.Consumed('>') {
				kind = FatArrow
			}

		case s.Match(',', ':', '(', ')', '[', ']', '{', '}'):
			kind = FromRepresentation(s.Advance())

			if kind == Illegal {
				panic("unreachable")
			}

		default:
			s.error(ErrIllegalCharacter, s.Pos())
			data = string(s.Advance())
		}

		if pos := s.Pos(); span.To != pos {
			span.To = text.PosFrom(pos.ID(), pos.Offset()-1)
		}

		if s.flags&SkipIllegal != 0 && kind == Illegal {
			return s.NextToken()
		}

		if span.From == span.To {
			span.To = text.NoPos
		}

		s.emitNewLine = kind != Semicolon && kind != Newline
		return Token{Kind: kind, Data: data, Span: span}
	}

	return Token{
		Kind: EOF,
		Span: text.Span{From: s.Pos()},
	}
}

func (s *Scanner) scanString() (data string, span text.Span, ok bool) {
	span = text.Span{From: s.Pos(), To: s.Pos()}
	quote := s.ExpectChar('"', '\'')
	buf := strings.Builder{}
	buf.WriteRune(quote)

	for {
		switch char := s.Peek(); char {
		case quote:
			span.To = s.Pos()
			s.Advance()
			buf.WriteRune(quote)
			data = buf.String()
			ok = true
			return

		case '\000', '\n', '\r':
			s.error(ErrUnterminatedStringLit, span.From)
			data = buf.String()
			span.To = s.Pos()
			return

		case '\\':
			backslashPos := s.Pos()

			switch char = s.Next(); char {
			case 'n':
				buf.WriteByte('\n')

			case 'r':
				buf.WriteByte('\r')

			case 't':
				buf.WriteByte('\t')

			case '\\':
				buf.WriteByte('\\')

			case '\'':
				buf.WriteByte('\'')

			case '"':
				buf.WriteByte('"')

			case 'x':
				if !s.parseBytes(&buf, 2) {
					// TODO error
				}

			case 'u':
				if !s.parseBytes(&buf, 4) {
					// TODO error
				}

			case 'U':
				if !s.parseBytes(&buf, 8) {
					// TODO error
				}

			default:
				s.error(ErrInvalidEscape, backslashPos)

				// NOTE not sure if invalid escape needs to be present in token
				buf.WriteByte('\\')
				buf.WriteRune(char)
			}

		case '$':
			// TODO interpolated string

		default:
			buf.WriteRune(s.Advance())
		}
	}
}

func (s *Scanner) scanNumber() (kind Kind, data string, span text.Span, ok bool) {
	buf := strings.Builder{}
	kind = Int
	span = text.Span{From: s.Pos(), To: s.Pos()}

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

			data = buf.String()
			ok = true
			return
		}

		s.parseNumber(nil, isZero, nil)

		if s.SkipWhile(isZero) > 0 {
			// TODO warning?
			s.error(ErrFirstDigitIsZero, span.From)
		}

		s.ConsumeFunc(isDigit)

		switch {
		case isDigit(s.Peek()):
			// TODO parse it like regular number
			s.error(ErrFirstDigitIsZero, span.From)
			kind = Illegal
			data = "0"
			return

		default:
			buf.WriteByte('0')
		}
	} else if !s.parseDecNumber(&buf) {
		kind = Illegal
		return
	}

	if s.Consumed('.') {
		buf.WriteByte('.')
		kind = Float

		if !s.parseNumber(&buf, isDigit, ErrExpectedDigitAfterPoint) {
			kind = Illegal
			return
		}
	}

	if s.Match('e', 'E') {
		buf.WriteRune(s.Advance())
		kind = Float

		if s.Match('+', '-') {
			buf.WriteRune(s.Advance())
		}

		if !s.parseDecNumber(&buf) {
			kind = Illegal
			return
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

	span.To = s.Pos()
	data = buf.String()
	ok = true
	return
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
