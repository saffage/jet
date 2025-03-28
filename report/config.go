package report

import (
	"io"
	"os"
)

// Specifies whether to output the actual code from the file.
var ShowCodeSnapshot = true

// Specifies whether to use unicode symbols in reports.
//
// Affects only punctuation used in generated report.
var UseUnicode = true

// Specifies a level of messages to be displayed.
var MinDisplayLevel = LevelTrace

// Specifies output.
var Output io.Writer = os.Stderr

var LineInfoStyle = LineInfoUnix

//go:generate stringer -type=Level -linecomment
type Level byte

const (
	LevelError   Level = iota // error
	LevelWarning              // warning
	LevelHint                 // hint
	LevelDebug                // debug
	LevelTrace                // trace
)

//go:generate stringer -type=LineInfo -linecomment
type LineInfo byte

const (
	LineInfoUnix LineInfo = iota // unix
)
