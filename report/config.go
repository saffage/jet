package report

import (
	"io"
	"os"
)

//go:generate stringer -type=Level -linecomment
type Level byte

const (
	LevelError   Level = iota // error
	LevelWarning              // warning
	LevelHint                 // hint
	LevelDebug                // debug
)

// Specifies whether colors will be used when printing messages.
var UseColors = true

// Specifies whether to output the actual code from the file.
var ShowCodeSnapshot = true

// Specifies a level of messages to be displayed.
var MinDisplayLevel = LevelHint

// Specifies output file.
var OutFile io.Writer = os.Stderr
