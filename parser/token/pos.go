package token

import (
	"strconv"
	"strings"

	"github.com/saffage/jet/config"
)

// Zero value is invalid location.
type Pos struct {
	FileID config.FileID
	Offset uint64
	Line   uint32
	Char   uint32
}

func (p Pos) WithEnd(end Pos) Range {
	return RangeFrom(p, end)
}

func (p Pos) WithStart(start Pos) Range {
	return RangeFrom(start, p)
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
func (l Pos) String() string {
	s := strings.Builder{}

	if file, present := config.Global.Files[l.FileID]; present && file.Path != "" {
		s.WriteString(file.Path)
		s.WriteByte(':')
	}

	if l.Line > 0 {
		s.WriteString(strconv.Itoa(int(l.Line)))

		if l.Char > 0 {
			s.WriteByte(':')
			s.WriteString(strconv.Itoa(int(l.Char)))
		}
	}

	if s.Len() == 0 {
		return "???"
	}

	return s.String()
}

func (l Pos) IsValid() bool {
	return l.FileID != 0 && l.Line > 0
}
