package report

import (
	"fmt"

	"github.com/saffage/jet/token"
)

// Specifies whether colors will be used when printing messages.
var UseColors = true

// Specifies whether to output the actual code from the file.
var ShowLine = true

// Specifies whether to output debug messages.
var IsDebug = false

// Reporter is an interface that is used to make the report prettier/clearer.
//
// Types implementing this interface must call the functions they need
// themselves (e.g. [TaggedErrorf]).
type Reporter interface {
	Report()
}

// Displays an error report for each of the errors.
//
// If the error implements the [Reporter] interface, it will be used instead
// of the usual Error() function.
func Errors(errors ...error) {
	for _, err := range errors {
		if reporter, ok := err.(Reporter); ok {
			reporter.Report()
		} else {
			Error(err.Error())
		}
	}
}

// Reports a message of the specified kind.
func Report(kind Kind, args ...any) {
	reportInternal(kind, "", fmt.Sprint(args...))
}

// [Debug] is a convenient helper function for reporting a
// debug message.
//
// Same as `Report(KindDebug, args...)`.
func Debug(format string, args ...any) {
	Report(KindDebug, args...)
}

// [Note] is a convenient helper function for reporting a
// note with a message.
//
// Same as `Report(KindNote, args...)`.
func Note(args ...any) {
	Report(KindNote, args...)
}

// [Hint] is a convenient helper function for reporting a
// hint with a message.
//
// Same as `Report(KindHint, args...)`.
func Hint(args ...any) {
	Report(KindHint, args...)
}

// [Warning] is a convenient helper function for reporting a
// warning with a message.
//
// Same as `Report(KindWarning, args...)`.
func Warning(args ...any) {
	Report(KindWarning, args...)
}

// [Error] is a convenient helper function for reporting an
// error with a formatted message.
//
// Same as `Report(KindError, args...)`.
func Error(format string, args ...any) {
	Report(KindError, args...)
}

// Reports a formatted message of the specified kind.
func Reportf(kind Kind, format string, args ...any) {
	reportInternal(kind, "", fmt.Sprintf(format, args...))
}

// [Debugf] is a convenient helper function for reporting a
// formatted debug message.
//
// Same as `Reportf(KindDebug, format, args...)`.
func Debugf(format string, args ...any) {
	Reportf(KindDebug, format, args...)
}

// [Notef] is a convenient helper function for reporting a
// note with a formatted message.
//
// Same as `Reportf(KindNote, format, args...)`.
func Notef(format string, args ...any) {
	Reportf(KindNote, format, args...)
}

// [Hintf] is a convenient helper function for reporting a
// hint with a formatted message.
//
// Same as `Reportf(KindHint, format, args...)`.
func Hintf(format string, args ...any) {
	Reportf(KindHint, format, args...)
}

// [Warningf] is a convenient helper function for reporting a
// warning with a formatted message.
//
// Same as `Reportf(KindWarning, format, args...)`.
func Warningf(format string, args ...any) {
	Reportf(KindWarning, format, args...)
}

// [Errorf] is a convenient helper function for reporting an
// error with a formatted message.
//
// Same as `Reportf(KindError, format, args...)`.
func Errorf(format string, args ...any) {
	Reportf(KindError, format, args...)
}

// Reports a tagged formatted message of the specified kind.
func TaggedReport(kind Kind, tag string, args ...any) {
	reportInternal(kind, tag, fmt.Sprint(args...))
}

// [TaggedDebug] is a convenient helper function for reporting a
// tagged debug message.
//
// Same as `TaggedReport(KindDebug, args...)`.
func TaggedDebug(tag string, args ...any) {
	TaggedReport(KindDebug, tag, args...)
}

// [TaggedNote] is a convenient helper function for reporting a
// note with a tagged message.
//
// Same as `TaggedReport(KindNote,, args...)`.
func TaggedNote(tag string, args ...any) {
	TaggedReport(KindNote, tag, args...)
}

// [TaggedHint] is a convenient helper function for reporting a
// hint with a tagged message.
//
// Same as `TaggedReport(KindHint, args...)`.
func TaggedHint(tag string, args ...any) {
	TaggedReport(KindHint, tag, args...)
}

// [TaggedWarning] is a convenient helper function for reporting a
// warning with a tagged message.
//
// Same as `TaggedReport(KindWarning, args...)`.
func TaggedWarning(tag string, args ...any) {
	TaggedReport(KindWarning, tag, args...)
}

// [TaggedError] is a convenient helper function for reporting an
// error with a tagged message.
//
// Same as `TaggedReport(KindError, args...)`.
func TaggedError(tag string, args ...any) {
	TaggedReport(KindError, tag, args...)
}

// Reports a tagged formatted message of the specified kind.
func TaggedReportf(kind Kind, tag, format string, args ...any) {
	reportInternal(kind, tag, fmt.Sprintf(format, args...))
}

// [TaggedDebugf] is a convenient helper function for reporting a
// tagged formatted debug message.
//
// Same as `TaggedReportf(KindDebug, format, args...)`.
func TaggedDebugf(tag, format string, args ...any) {
	TaggedReportf(KindDebug, tag, format, args...)
}

// [TaggedNotef] is a convenient helper function for reporting a
// note with a tagged formatted message.
//
// Same as `TaggedReportf(KindNote, format, args...)`.
func TaggedNotef(tag, format string, args ...any) {
	TaggedReportf(KindNote, tag, format, args...)
}

// [TaggedHintf] is a convenient helper function for reporting a
// hint with a tagged formatted message.
//
// Same as `TaggedReportf(KindHint, format, args...)`.
func TaggedHintf(tag, format string, args ...any) {
	TaggedReportf(KindHint, tag, format, args...)
}

// [TaggedWarningf] is a convenient helper function for reporting a
// warning with a tagged formatted message.
//
// Same as `TaggedReportf(KindWarning, format, args...)`.
func TaggedWarningf(tag, format string, args ...any) {
	TaggedReportf(KindWarning, tag, format, args...)
}

// [TaggedErrorf] is a convenient helper function for reporting an
// error with a tagged formatted message.
//
// Same as `TaggedReportf(KindError, format, args...)`.
func TaggedErrorf(tag, format string, args ...any) {
	TaggedReportf(KindError, tag, format, args...)
}

// [ReportAt] function reports a message of the specified kind,
// highlighting the specified range with an underscore.
//
// NOTE: this function uses [config.global] to get information about the file
// and also its buffer.
func ReportAt(kind Kind, start, end token.Pos, args ...any) {
	reportAtInternal(kind, "", start, end, fmt.Sprint(args...))
}

// [DebugAt] is a convenient helper function for reporting a
// debug message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindDebug, message, start, end)`.
func DebugAt(start, end token.Pos, args ...any) {
	ReportAt(KindDebug, start, end, args...)
}

// [NoteAt] is a convenient helper function for reporting a
// note with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindNote, message, start, end)`.
func NoteAt(start, end token.Pos, args ...any) {
	ReportAt(KindNote, start, end, args...)
}

// [HintAt] is a convenient helper function for reporting a
// hint with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindHint, message, start, end)`.
func HintAt(start, end token.Pos, args ...any) {
	ReportAt(KindHint, start, end, args...)
}

// [WarningAt] is a convenient helper function for reporting a
// warning with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindWarning, message, start, end)`.
func WarningAt(start, end token.Pos, args ...any) {
	ReportAt(KindWarning, start, end, args...)
}

// [ErrorAt] is a convenient helper function for reporting an
// error with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindError, message, start, end)`.
func ErrorAt(start, end token.Pos, args ...any) {
	ReportAt(KindError, start, end, args...)
}

// [TaggedReportAt] function reports a tagged message of the specified kind,
// highlighting the specified range with an underscore.
//
// NOTE: this function uses [config.global] to get information about the file
// and also its buffer.
func TaggedReportAt(kind Kind, tag string, start, end token.Pos, args ...any) {
	reportAtInternal(kind, tag, start, end, fmt.Sprint(args...))
}

// [TaggedDebugAt] is a convenient helper function for reporting a
// tagged debug message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindDebug, tag, message, start, end)`.
func TaggedDebugAt(tag string, start, end token.Pos, args ...any) {
	TaggedReportAt(KindDebug, tag, start, end, args...)
}

// [TaggedNoteAt] is a convenient helper function for reporting a
// note with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindNote, tag, message, start, end)`.
func TaggedNoteAt(tag string, start, end token.Pos, args ...any) {
	TaggedReportAt(KindNote, tag, start, end, args...)
}

// [TaggedHintAt] is a convenient helper function for reporting a
// hint with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindHint, tag, message, start, end)`.
func TaggedHintAt(tag string, start, end token.Pos, args ...any) {
	TaggedReportAt(KindHint, tag, start, end, args...)
}

// [TaggedWarningAt] is a convenient helper function for reporting a
// warning with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindWarning, tag, message, start, end)`.
func TaggedWarningAt(tag string, start, end token.Pos, args ...any) {
	TaggedReportAt(KindWarning, tag, start, end, args...)
}

// [TaggedErrorAt] is a convenient helper function for reporting an
// error with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindError, tag, message, start, end)`.
func TaggedErrorAt(tag string, start, end token.Pos, args ...any) {
	TaggedReportAt(KindError, tag, start, end, args...)
}
