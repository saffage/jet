package text

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidPos         = errors.New("invalid pos")
	ErrInvalidPosEncoded  = errors.New("invalid pos value encoded")
	ErrInvalidSpanEncoded = errors.New("invalid span value encoded")
	ErrFileIDOutOfRange   = errors.New("position file ID value out of range")
	ErrOffsetOutOfRange   = errors.New("position offset value out of range")
)

const NoPos Pos = 0

type Position struct {
	Filepath string
	ID       FileID
	Offset   int
	Line     int
	Char     int
}

func (p Position) IsValid() bool {
	// 0 offset is allowed
	return p.Filepath != "" && p.ID != 0 && p.Line != 0 && p.Char != 0
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Filepath, p.Line, p.Char)
}

type Pos uint64

func (p Pos) ID() FileID                { return FileID(p & fileid_mask) }
func (p Pos) Offset() int               { return int((p >> fileid_bits) & offset_mask) }
func (p Pos) IsValid() bool             { return p != NoPos }
func (p Pos) WithOffset(offset int) Pos { return PosFrom(p.ID(), p.Offset()+offset) }
func (p Pos) Renderable() bool          { return p.IsValid() }

func (p Pos) String() string {
	text, err := p.MarshalText()

	if err != nil {
		return "<invalid-pos>"
	}

	return string(text)
}

func (p Pos) Render(buf Writer) {
	if !p.IsValid() {
		panic("invalid pos")
	}

	if buf, _ := buf.(*strings.Builder); buf != nil {
		buf.Grow(10)
	}

	buf.WriteString(strconv.Itoa(int(p.ID())))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(p.Offset()))
}

func (p Pos) MarshalText() ([]byte, error) {
	if !p.IsValid() {
		return nil, ErrInvalidPos
	}

	buf := strings.Builder{}

	p.Render(&buf)

	return []byte(buf.String()), nil
}

func (p *Pos) UnmarshalText(text []byte) error {
	idString, offsetString, ok := strings.Cut(string(text), ",")

	if !ok {
		return ErrInvalidPosEncoded
	}

	offset, err := strconv.Atoi(idString)

	if err != nil {
		return err
	}
	if offset > offset_mask {
		return ErrOffsetOutOfRange
	}

	id, err := strconv.Atoi(offsetString)

	if err != nil {
		return err
	}
	if id > fileid_mask {
		return ErrFileIDOutOfRange
	}

	if id == 0 {
		*p = NoPos
	} else {
		*p = PosFrom(FileID(id), offset)
	}

	return nil
}

func PosFrom(id FileID, offset int) Pos {
	if offset < 0 {
		panic("negative offset")
	}
	if offset > offset_mask {
		panic("offset overflow")
	}
	if id > fileid_mask {
		panic("fileid overflow")
	}
	if id <= 0 {
		panic("invalid fileid")
	}
	return Pos(
		uint64(id)&fileid_mask |
			(uint64(offset)&offset_mask)<<fileid_bits)
}

const (
	_ Pos = 1<<(fileid_bits+offset_bits) - 1

	fileid_bits = 16
	offset_bits = 48

	fileid_mask = 1<<fileid_bits - 1
	offset_mask = 1<<offset_bits - 1
)

type Span struct {
	From Pos
	To   Pos
}

func (s Span) ID() FileID    { return s.From.ID() }
func (s Span) IsPos() bool   { return !s.To.IsValid() || s.To == s.From }
func (s Span) IsValid() bool { return s.From.IsValid() }

func (s Span) Renderable() bool { return s.IsValid() }

func (s Span) Render(buf Writer) {
	if !s.IsValid() {
		panic("invalid pos")
	}

	if buf, _ := buf.(*strings.Builder); buf != nil {
		buf.Grow(16)
	}

	buf.WriteString(strconv.Itoa(int(s.ID())))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(s.From.Offset()))

	if !s.IsPos() {
		buf.WriteByte(',')
		buf.WriteString(strconv.Itoa(s.To.Offset()))
	}
}

func (s Span) String() string {
	text, err := s.MarshalText()

	if err != nil {
		return "<invalid-span>"
	}

	return string(text)
}

func (s Span) MarshalText() ([]byte, error) {
	if !s.IsValid() {
		return nil, ErrInvalidPos
	}

	buf := strings.Builder{}

	s.Render(&buf)

	return []byte(buf.String()), nil
}

func (s *Span) UnmarshalText(text []byte) error {
	components := strings.SplitN(string(text), ",", 3)

	var id, from, to int
	var err error

	switch len(components) {
	case 3:
		to, err = strconv.Atoi(components[2])

		if err != nil {
			return err
		}
		if to > offset_mask {
			return ErrOffsetOutOfRange
		}

		fallthrough

	case 2:
		from, err = strconv.Atoi(components[1])

		if err != nil {
			return err
		}
		if from > offset_mask {
			return ErrOffsetOutOfRange
		}

		id, err = strconv.Atoi(components[0])

		if err != nil {
			return err
		}
		if id > fileid_mask {
			return ErrFileIDOutOfRange
		}

	default:
		return ErrInvalidSpanEncoded
	}

	if id == 0 {
		*s = Span{}
	} else {
		s.From = PosFrom(FileID(id), from)
		s.To = PosFrom(FileID(id), to)
	}

	return nil
}
