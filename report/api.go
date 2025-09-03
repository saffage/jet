// Package report implements compiler problem reporting system.
package report

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/saffage/jet/text"
)

// Informer is an interface used to inform reporter how to report a problem.
type Informer interface {
	Info() *Info
}

// Render renders the error for a user, writing it into the [Output] variable.
//
// If the error implements the [Renderer] or [Informer] interfaces, it will be
// used instead of the usual [Error] method.
//
// Note that errors joined using [errors.Join] will not be shown as separate
// errors, use [Join] instead.
func Render(err error) (rendered bool) {
	return RenderTo(Output, err)
}

// RenderTo renders the error for a user, writing it into the specified
// writer.
//
// If the error implements [text.Renderer] or [Informer] interfaces,
// it will be used instead of the usual [Error] method.
//
// Note that errors joined using [errors.Join] will not be shown as separate
// errors, use [Join] instead.
func RenderTo(w io.Writer, err error) (rendered bool) {
	buf := bufio.NewWriter(w)
	rendered = render(buf, err)
	return rendered
}

func Debug(format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	report(
		LevelDebug,
		fmt.Sprintf("%s:%d", file, line),
		fmt.Sprintf(format, args...),
	)
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

	const labelBufferSize = 10

	buf := strings.Builder{}
	buf.Grow(len(message) + len(tag) + labelBufferSize)
	writeLabel(&buf, level, tag)
	titleStyle.Fprint(&buf, message)

	fmt.Fprintln(Output, buf.String())
}

func render(buf text.Writer, err error) (rendered bool) {
	switch err := err.(type) {
	case nil:
		// Ignore

	case text.Renderer:
		if rendered = err.Render(buf); !rendered {
			rendered = renderUnwrappedError(buf, err)
		}

	case Informer:
		if info := err.Info(); info != nil {
			info.Render(buf)
			rendered = true
		} else {
			rendered = renderUnwrappedError(buf, err)
		}

	default:
		info := Info{Title: err.Error()}
		info.Render(buf)
		rendered = true
	}
	return rendered
}

func renderUnwrappedError(buf text.Writer, err any) (rendered bool) {
	switch err := err.(type) {
	case interface{ Unwrap() error }:
		rendered = render(buf, err.Unwrap())

	case interface{ Unwrap() []error }:
		for _, err := range err.Unwrap() {
			rendered = render(buf, err)
		}
	}
	return rendered
}
