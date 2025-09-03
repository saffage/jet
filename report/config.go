package report

//go:generate go tool stringer -type=Level -linecomment

import (
	"io"
	"os"
)

// ShowCodeSnapshot specifies whether to output the actual code from the file.
var ShowCodeSnapshot = true

// UseUnicode specifies whether to use unicode symbols in reports.
//
// Affects only punctuation used in generated report.
var UseUnicode = true

// MinDisplayLevel specifies a level of messages to be displayed.
var MinDisplayLevel = LevelTrace

// Output specifies the output file\stream.
var Output io.Writer = os.Stderr

type Level byte

const (
	LevelError   Level = iota // error
	LevelWarning              // warning
	LevelHint                 // hint
	LevelDebug                // debug
	LevelTrace                // trace
)
