package report

import (
	"fmt"
	"runtime"
	"strings"
)

// Renderer is an interface for rendering content into a buffer.
//
// There are also a method to check if the content is renderable and can
// perform the rendering without any errors.
type Renderer interface {
	Render(buf *strings.Builder)
	Renderable() bool
}

// Informer is an interface used to inform reporter how to report a problem.
type Informer interface {
	Info() *Info
}

// Render renders the error for a user, writing it into the specified [Output].
//
// If the error implements the [Renderer] or [Informer] interfaces, it will be
// used instead of the usual [Error] method.
//
// Note that errors joined using [errors.Join] will not be shown as separate
// errors, use [Join] instead.
func Render(err error) {
	buf := strings.Builder{}
	render(&buf, err)
	Output.Write([]byte(buf.String()))
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

func render(buf *strings.Builder, err error) {
	switch err := err.(type) {
	case nil:
		// Ignore

	case Renderer:
		if err.Renderable() {
			err.Render(buf)
		} else {
			renderUnwrappedError(buf, err)
		}

	case Informer:
		if info := err.Info(); info != nil {
			info.Render(buf)
		} else {
			renderUnwrappedError(buf, err)
		}

	default:
		info := Info{Title: err.Error()}
		info.Render(buf)
	}
}

func renderUnwrappedError(buf *strings.Builder, err any) {
	switch err := err.(type) {
	case interface{ Unwrap() error }:
		render(buf, err.Unwrap())

	case interface{ Unwrap() []error }:
		for _, err := range err.Unwrap() {
			render(buf, err)
		}
	}
}
