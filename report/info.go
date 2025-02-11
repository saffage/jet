package report

import (
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

// Informer is an interface used to inform reporter how to report a problem.
type Informer interface {
	Info() *Info
}

type Info struct {
	Tag            string
	Title          string
	SelectionHint  string
	Hints          []HintInfo
	SelectionRange token.Range
	Level          Level
}

type HintInfo struct {
	Message         string
	Suggestion      string // if empty, then config will be used
	SuggestionRange token.Range
}

func (info *Info) Error() string {
	buf := strings.Builder{}

	if info.Tag != "" {
		buf.WriteString(info.Tag)
		buf.WriteByte(' ')
		buf.WriteString(info.Level.String())
		buf.WriteString(": ")
	}

	if strings.TrimSpace(info.Title) == "" {
		buf.WriteString(emptyMessage)
	} else {
		buf.WriteString(info.Title)
	}

	if info.SelectionHint != "" {
		buf.WriteString(" (")
		buf.WriteString(info.SelectionHint)
		buf.WriteString(")")
	}

	return buf.String()
}

func (info *Info) Report() {
	if info.Level > MinDisplayLevel {
		return
	}

	// We do it here because the message will not be empty if range is specified.
	if strings.TrimSpace(info.Title) == "" {
		info.Title = emptyMessage
	}

	var file *text.File

	if info.SelectionRange.IsValid() {
		file = config.File(info.SelectionRange.FileID)

		if file == nil {
			panic("unreachable")
		}
	}

	buf := strings.Builder{}

	if info.SelectionRange.IsValid() {
		snapshot := genCodeSnapshot(
			info.Level,
			info.SelectionHint,
			info.SelectionRange,
			file,
		)
		buf.WriteByte('\n')
		buf.WriteString(snapshot)
	} else if info.SelectionHint != "" {
		buf.WriteString(" (")
		buf.WriteString(info.SelectionHint)
		buf.WriteByte(')')
	}

	buf.WriteString(genHint(info.Hints, info.SelectionRange, file))
	report(info.Level, info.Tag, info.Title+buf.String())
}

func (info *Info) Info() *Info { return info }

func (info *Info) With(hint HintInfo) *Info {
	info.Hints = append(info.Hints, hint)
	return info
}
