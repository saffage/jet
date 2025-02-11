package token

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/text"
)

// Zero value is invalid location.
type Pos struct {
	FileID text.FileID
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
		s.WriteString(config.File(pos.FileID).Path + ":")
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
	return Range{FileID: pos.FileID, Start: rangePos{Line: pos.Line, Char: pos.Char}, End: rangePos{Line: end.Line, Char: end.Char}}
}

func (pos Pos) WithStart(start Pos) Range {
	return Range{FileID: start.FileID, Start: rangePos{Line: start.Line, Char: start.Char}, End: rangePos{Line: pos.Line, Char: pos.Char}}
}

func (pos Pos) WithLen(i uint32) Range {
	end := Pos{
		FileID: pos.FileID,
		Line:   pos.Line,
		Char:   pos.Char + i,
	}
	return Range{FileID: pos.FileID, Start: rangePos{Line: pos.Line, Char: pos.Char}, End: rangePos{Line: end.Line, Char: end.Char}}
}

func (pos Pos) AsRange() Range {
	return Range{FileID: pos.FileID, Start: rangePos{Line: pos.Line, Char: pos.Char}, End: rangePos{Line: pos.Line, Char: pos.Char}}
}
