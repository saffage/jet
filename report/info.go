package report

import (
	"strings"

	"github.com/saffage/jet/config"
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
	Message    string
	Suggestion string
	HintRange  token.Range
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

func (info *Info) Report(cfg *config.Config) {
	if info.Level > MinDisplayLevel {
		return
	}

	// We do it here because the message will not be empty if range is specified.
	if strings.TrimSpace(info.Title) == "" {
		info.Title = emptyMessage
	}

	var file *config.File

	if info.SelectionRange.IsValid() {
		file = cfg.Files[info.SelectionRange.FileID]

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
			cfg,
		)
		buf.WriteByte('\n')
		buf.WriteString(snapshot)
	} else if info.SelectionHint != "" {
		buf.WriteString(" (")
		buf.WriteString(info.SelectionHint)
		buf.WriteByte(')')
	}

	if info.SelectionRange.IsValid() {
		buf.WriteString(genHint(info.Hints, info.SelectionRange, file, cfg))
	}

	if UseColors {
		report(info.Level, info.Tag, info.Title+buf.String())
	} else {
		report(info.Level, info.Tag, info.Title+buf.String())
	}
}

func (info *Info) With(hint HintInfo) *Info {
	info.Hints = append(info.Hints, hint)
	return info
}
