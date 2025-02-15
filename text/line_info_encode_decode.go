package text

import (
	"errors"
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

func (s Span) String() string {
	text, err := s.MarshalText()

	if err != nil {
		return "<invalid-span>"
	}

	return string(text)
}

func (s Span) Render(buf *strings.Builder) error {
	if !s.IsValid() {
		return ErrInvalidPos
	}

	buf.Grow(16)
	buf.WriteString(strconv.Itoa(int(s.ID())))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(s.From.Offset()))

	if !s.IsPos() {
		buf.WriteByte(',')
		buf.WriteString(strconv.Itoa(s.To.Offset()))
	}

	return nil
}

func (s Span) MarshalText() ([]byte, error) {
	buf := strings.Builder{}
	err := s.Render(&buf)
	return []byte(buf.String()), err
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

func (p Pos) String() string {
	text, err := p.MarshalText()

	if err != nil {
		return "<invalid-pos>"
	}

	return string(text)
}

func (p Pos) Render(buf *strings.Builder) error {
	if !p.IsValid() {
		return ErrInvalidPos
	}

	buf.Grow(10)
	buf.WriteString(strconv.Itoa(int(p.ID())))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(p.Offset()))
	return nil
}

func (p Pos) MarshalText() ([]byte, error) {
	buf := strings.Builder{}
	err := p.Render(&buf)
	return []byte(buf.String()), err
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
