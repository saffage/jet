package token

import (
	"fmt"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/text"
)

type rangePos struct {
	Line uint32
	Char uint32
}

type Range struct {
	FileID     text.FileID
	Start, End struct {
		Line uint32
		Char uint32
	}
}

func (rng Range) StartPos() Pos {
	return Pos{
		FileID: rng.FileID,
		Line:   rng.Start.Line,
		Char:   rng.Start.Char,
	}
}

func (rng Range) EndPos() Pos {
	return Pos{
		FileID: rng.FileID,
		Line:   rng.End.Line,
		Char:   rng.End.Char,
	}
}

func (rng Range) String() string {
	filepath, start, end := "", "", ""

	if rng.FileID != 0 {
		filepath = config.File(rng.FileID).Path
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
	return rng.FileID != 0 && rng.Start.Line != 0
}
