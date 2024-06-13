package token

import (
	"fmt"
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

	if fileinfo, present := config.Global.Files[l.FileID]; present && fileinfo.Path != "" {
		s.WriteString(fileinfo.Path + ":")
	}

	if l.Line > 0 {
		s.WriteString(fmt.Sprintf("%d", l.Line))

		if l.Char > 0 {
			s.WriteString(fmt.Sprintf(":%d", l.Char))
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
