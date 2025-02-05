package report

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/saffage/jet/config"
)

// Specifies whether colors will be used when printing messages.
var UseColors = true

// Specifies whether to output the actual code from the file.
var ShowCodeSnapshot = true

// Specifies a level of messages to be displayed.
var MinDisplayLevel = LevelHint

// Specifies output file.
var OutFile io.Writer = os.Stderr

// If the error implements the [Informer] interface, it will be used instead
// of the usual [Error] function.
func Report(cfg *config.Config, errs ...error) {
	for _, err := range errs {
		switch err := err.(type) {
		case nil:

		case Informer:
			err.Info().Report(cfg)

		case interface{ Unwrap() []error }:
			Report(cfg, err.Unwrap()...)

		case interface{ Unwrap() error }:
			Report(cfg, err.Unwrap())

		default:
			(&Info{Title: err.Error()}).Report(cfg)
		}
	}
}

func Debug(format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	report(LevelDebug, fmt.Sprintf("%s:%d", file, line), fmt.Sprintf(format, args...))
}

func DebugX(tag, format string, args ...any) {
	report(LevelDebug, tag, fmt.Sprintf(format, args...))
}

func Hint(format string, args ...any) {
	report(LevelHint, "", fmt.Sprintf(format, args...))
}

func HintX(tag, format string, args ...any) {
	report(LevelHint, tag, fmt.Sprintf(format, args...))
}

func Warning(format string, args ...any) {
	report(LevelWarning, "", fmt.Sprintf(format, args...))
}

func WarningX(tag string, format string, args ...any) {
	report(LevelWarning, tag, fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	report(LevelError, "", fmt.Sprintf(format, args...))
}

func ErrorX(tag string, format string, args ...any) {
	report(LevelError, tag, fmt.Sprintf(format, args...))
}

const emptyMessage = "{no message provided}"

func report(level Level, tag, message string) {
	if level > MinDisplayLevel {
		return
	}

	if strings.TrimSpace(message) == "" {
		message = emptyMessage
	}

	fmt.Fprintln(OutFile, level.Label(tag), message)
}
