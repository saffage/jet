package report

import (
	"fmt"
	"runtime"

	"github.com/saffage/jet/parser/token"
)

// Specifies whether colors will be used when printing messages.
var UseColors = true

// Specifies whether to output the actual code from the file.
var ShowCodeSnapshot = true

// Specifies a level of messages to be displayed.
var MinDisplayLevel = LevelHint

// Problem is an interface that is used to make a [Info].
type Problem interface {
	Info() *Info
}

// Displays an problem report for each of the problems.
//
// If the error implements the [Problem] interface, it will be used instead
// of the usual Error() function.
func Error(errs ...error) {
	for _, err := range errs {
		switch err := err.(type) {
		case nil:
			// Do nothing

		case Problem:
			err.Info().Report()

		case interface{ Unwrap() error }:
			Error(err.Unwrap())

		case interface{ Unwrap() []error }:
			Error(err.Unwrap()...)

		default:
			p := Info{Kind: KindError, Title: err.Error()}
			p.Report()
		}
	}
}

func Debug(format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	report(LevelDebug, fmt.Sprintf("%s:%d", file, line), fmt.Sprintf(format, args...))
}

func Hint(format string, args ...any) {
	report(LevelHint, "", fmt.Sprintf(format, args...))
}

func WarningX(tag string, format string, args ...any) {
	report(LevelWarning, tag, fmt.Sprintf(format, args...))
}

func ErrorX(tag string, format string, args ...any) {
	report(LevelError, tag, fmt.Sprintf(format, args...))
}

func DebugX(tag, format string, args ...any) {
	report(LevelDebug, tag, fmt.Sprintf(format, args...))
}

func HintX(tag, format string, args ...any) {
	report(LevelHint, tag, fmt.Sprintf(format, args...))
}

func WarningRangeX(tag string, rng token.Range, format string, args ...any) {
	info := Info{
		Tag:   tag,
		Title: fmt.Sprintf(format, args...),
		Range: rng,
	}
	info.reportWithCodeSnapshot(LevelWarning)
}

func ErrorRangeX(tag string, rng token.Range, format string, args ...any) {
	info := Info{
		Tag:   tag,
		Title: fmt.Sprintf(format, args...),
		Range: rng,
	}
	info.reportWithCodeSnapshot(LevelError)
}
