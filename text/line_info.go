package text

const NoPos Pos = 0

type Position struct {
	Pos
	Path string
	Line int
	Char int
}

func (p *Position) IsValid() bool { return p != nil && *p != Position{} }

type Pos uint64

func (p Pos) ID() FileID    { return FileID(p & fileid_mask) }
func (p Pos) Offset() int   { return int((p >> fileid_bits) & offset_mask) }
func (p Pos) IsValid() bool { return p != 0 }

type Span struct {
	From Pos
	To   Pos
}

func (s Span) ID() FileID    { return s.From.ID() }
func (s Span) IsPos() bool   { return s.From == s.To }
func (s Span) IsValid() bool { return s.From.IsValid() && s.To.IsValid() }

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
