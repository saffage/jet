package token

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/config"
)

// Zero value is invalid location.
type Pos struct {
	FileID config.FileID
	Line   uint32
	Char   uint32
}

// Uses [config.Global] to get the file info.
//
// Return string in one of this formats depending on location data:
//   - "file"
//   - "file:line"
//   - "file:line:column"
//   - "line"
//   - "line:column"
//   - "???"
func (pos Pos) String() string {
	s := strings.Builder{}

	if pos.FileID != 0 {
		s.WriteString(config.Global.Files[pos.FileID].Path + ":")
	}

	if pos.Line > 0 {
		s.WriteString(fmt.Sprintf("%d", pos.Line))

		if pos.Char > 0 {
			s.WriteString(fmt.Sprintf(":%d", pos.Char))
		}
	}

	if s.Len() == 0 {
		return "???"
	}

	return s.String()
}

func (pos Pos) IsValid() bool {
	return pos.FileID != 0 && pos.Line > 0
}

func (pos Pos) WithEnd(end Pos) Range {
	return RangeFrom(pos, end)
}

func (pos Pos) WithStart(start Pos) Range {
	return RangeFrom(start, pos)
}

func (pos Pos) WithLen(i uint32) Range {
	return RangeFrom(pos, Pos{
		FileID: pos.FileID,
		Line:   pos.Line,
		Char:   pos.Char + i,
	})
}

func (pos Pos) AsRange() Range {
	return RangeFrom(pos, pos)
}
