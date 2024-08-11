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

type Range struct {
	FileID config.FileID
	Start  RangePos
	End    RangePos
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

	if fileinfo, ok := config.Global.Files[rng.FileID]; ok &&
		fileinfo.Path != "" {
		filepath = fileinfo.Path
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

	return fmt.Sprintf("%s:%s..%s", filepath, start, end)
}

func (rng Range) IsValid() bool {
	return rng.FileID != 0 && rng.Start.Line != 0 && rng.Start.Offset != 0
}
