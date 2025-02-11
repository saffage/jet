package report

import (
	"fmt"
	"runtime"
	"strings"
)

// If the error implements the [Informer] interface, it will be used instead
// of the usual [Error] function.
//
// Note that errors joined using [errors.Join] will not be shown as separate
// errors.
func Report(errs ...error) {
	for _, err := range errs {
		switch err := err.(type) {
		case nil:
			// Ignore

		case Informer:
			if info := err.Info(); info != nil {
				info.Report()
			}

			switch err := err.(type) {
			case interface{ Unwrap() error }:
				Report(err.Unwrap())

			case interface{ Unwrap() []error }:
				Report(err.Unwrap()...)
			}

		default:
			info := Info{Title: err.Error()}
			info.Report()
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
