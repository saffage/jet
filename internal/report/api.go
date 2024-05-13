package report

import "github.com/saffage/jet/token"

// Specifies whether colors will be used when printing messages.
var UseColors = true

// Specifies whether to output the actual code from the file.
var ShowLine = true

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
// of the usual error report.
func Report(errors ...error) {
	for _, err := range errors {
		if reporter, ok := err.(Reporter); ok {
			reporter.Report()
		} else {
			Errorf(err.Error())
		}
	}
}

// Reports a formatted message of the specified kind.
func Reportf(kind Kind, format string, args ...any) {
	reportfInternal(kind, "", format, args...)
}

// Reports a tagged formatted message of the specified kind.
func TaggedReportf(kind Kind, tag, format string, args ...any) {
	reportfInternal(kind, tag, format, args...)
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

// [InternalErrorf] is a convenient helper function for reporting an
// internal error with a formatted message.
//
// Same as `Reportf(KindInternalError, format, args...)`.
func InternalErrorf(format string, args ...any) {
	Reportf(KindInternalError, format, args...)
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

// [TaggedInternalErrorf] is a convenient helper function for reporting an
// internal error with a tagged formatted message.
//
// Same as `TaggedReportf(KindInternalError, format, args...)`.
func TaggedInternalErrorf(tag, format string, args ...any) {
	TaggedReportf(KindInternalError, tag, format, args...)
}

// [ReportAt] function reports a message of the specified kind,
// highlighting the specified range with an underscore.
//
// NOTE: this function uses [config.global] to get information about the file
// and also its buffer.
func ReportAt(kind Kind, message string, start, end token.Loc) {
	reportAtInternal(kind, "", message, start, end)
}

// [TaggedReportAt] function reports a tagged message of the specified kind,
// highlighting the specified range with an underscore.
//
// NOTE: this function uses [config.global] to get information about the file
// and also its buffer.
func TaggedReportAt(kind Kind, tag, message string, start, end token.Loc) {
	reportAtInternal(kind, tag, message, start, end)
}

// [NoteAt] is a convenient helper function for reporting a
// note with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindNote, message, start, end)`.
func NoteAt(message string, start, end token.Loc) {
	ReportAt(KindNote, message, start, end)
}

// [HintAt] is a convenient helper function for reporting a
// hint with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindHint, message, start, end)`.
func HintAt(message string, start, end token.Loc) {
	ReportAt(KindHint, message, start, end)
}

// [WarningAt] is a convenient helper function for reporting a
// warning with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindWarning, message, start, end)`.
func WarningAt(message string, start, end token.Loc) {
	ReportAt(KindWarning, message, start, end)
}

// [ErrorAt] is a convenient helper function for reporting an
// error with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindError, message, start, end)`.
func ErrorAt(message string, start, end token.Loc) {
	ReportAt(KindError, message, start, end)
}

// [InternalErrorAt] is a convenient helper function for reporting an
// internal error with a message, highlighting the specified range
// with an underscore.
//
// Same as `ReportAt(KindInternalError, message, start, end)`.
func InternalErrorAt(message string, start, end token.Loc) {
	ReportAt(KindInternalError, message, start, end)
}

// [TaggedNoteAt] is a convenient helper function for reporting a
// note with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindNote, tag, message, start, end)`.
func TaggedNoteAt(tag, message string, start, end token.Loc) {
	TaggedReportAt(KindNote, tag, message, start, end)
}

// [TaggedHintAt] is a convenient helper function for reporting a
// hint with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindHint, tag, message, start, end)`.
func TaggedHintAt(tag, message string, start, end token.Loc) {
	TaggedReportAt(KindHint, tag, message, start, end)
}

// [TaggedWarningAt] is a convenient helper function for reporting a
// warning with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindWarning, tag, message, start, end)`.
func TaggedWarningAt(tag, message string, start, end token.Loc) {
	TaggedReportAt(KindWarning, tag, message, start, end)
}

// [TaggedErrorAt] is a convenient helper function for reporting an
// error with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindError, tag, message, start, end)`.
func TaggedErrorAt(tag, message string, start, end token.Loc) {
	TaggedReportAt(KindError, tag, message, start, end)
}

// [TaggedInternalErrorAt] is a convenient helper function for reporting an
// internal error with a tagged message, highlighting the specified range
// with an underscore.
//
// Same as `TaggedReportAt(KindInternalError, tag, message, start, end)`.
func TaggedInternalErrorAt(tag, message string, start, end token.Loc) {
	TaggedReportAt(KindInternalError, tag, message, start, end)
}
