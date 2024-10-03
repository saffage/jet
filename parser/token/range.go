package token

import (
	"fmt"

	"github.com/saffage/jet/config"
)

type RangePos struct {
	Offset uint64
	Line   uint32
	Char   uint32
}

func (pos RangePos) IsValid() bool {
	return pos.Line != 0 && pos.Offset != 0
}

type Range struct {
	FileID config.FileID
	Start  RangePos
	End    RangePos
}

func RangeFrom(start, end Pos) Range {
	if start.FileID != end.FileID {
		panic(fmt.Sprintf(
			"start & end position have different file IDs (%d and %d)",
			start.FileID,
			end.FileID,
		))
	}

	return Range{
		FileID: start.FileID,
		Start: RangePos{
			Offset: start.Offset,
			Line:   start.Line,
			Char:   start.Char,
		},
		End: RangePos{
			Offset: end.Offset,
			Line:   end.Line,
			Char:   end.Char,
		},
	}
}

func (rng Range) StartPos() Pos {
	return Pos{
		FileID: rng.FileID,
		Offset: rng.Start.Offset,
		Line:   rng.Start.Line,
		Char:   rng.Start.Char,
	}
}

func (rng Range) EndPos() Pos {
	return Pos{
		FileID: rng.FileID,
		Offset: rng.End.Offset,
		Line:   rng.End.Line,
		Char:   rng.End.Char,
	}
}

func (rng Range) String() string {
	filepath, start, end := "", "", ""

	if file, present := config.Global.Files[rng.FileID]; present && file.Path != "" {
		filepath = file.Path
	}

	if rng.Start.Line > 0 {
		if rng.Start.Char > 0 {
			start = fmt.Sprintf("%d:%d", rng.Start.Line, rng.Start.Char)
		} else {
			start = fmt.Sprintf("%d", rng.Start.Line)
		}
	}

	if rng.End.Line > 0 {
		if rng.End.Char > 0 {
			end = fmt.Sprintf("%d:%d", rng.End.Line, rng.End.Char)
		} else {
			end = fmt.Sprintf("%d", rng.End.Line)
		}
	}

	if filepath == "" && start == "" && end == "" {
		return "???"
	}

	return fmt.Sprintf("%s:%s", filepath, start)
}

func (rng Range) IsValid() bool {
	return rng.FileID != 0 && rng.Start.Line != 0 && rng.Start.Offset != 0
}
