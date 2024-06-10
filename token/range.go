package token

import "github.com/saffage/jet/config"

type Range struct {
	FileID     config.FileID
	Start, End struct {
		Offset uint64
		Line   uint32
		Char   uint32
	}
}

func (rng Range) StartLoc() Pos {
	return Pos{
		FileID: rng.FileID,
		Offset: rng.Start.Offset,
		Line:   rng.Start.Line,
		Char:   rng.Start.Char,
	}
}

func (rng Range) EndLoc() Pos {
	return Pos{
		FileID: rng.FileID,
		Offset: rng.End.Offset,
		Line:   rng.End.Line,
		Char:   rng.End.Char,
	}
}

func (rng Range) String() string {
	return rng.StartLoc().String()
}

func (rng Range) IsValid() bool {
	return rng.FileID != 0 && rng.Start.Line != 0 && rng.Start.Offset != 0
}
