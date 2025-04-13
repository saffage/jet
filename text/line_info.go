package text

import "fmt"

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

type Span struct {
	From Pos
	To   Pos
}

func (s Span) ID() FileID    { return s.From.ID() }
func (s Span) IsPos() bool   { return !s.To.IsValid() || s.To == s.From }
func (s Span) IsValid() bool { return s.From.IsValid() }

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
