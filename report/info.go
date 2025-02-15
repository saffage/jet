package report

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/text"
)

const initialReportBufferSize = 512

// Informer is an interface used to inform reporter how to report a problem.
type Informer interface {
	Info() *Info
}

type Info struct {
	Title       string
	Tag         string
	Suggestions []Suggestion
	Selection   Selection
	Level       Level
}

type Selection struct {
	CustomContent string
	Hint          string
	Range         text.Span
}

type Suggestion struct {
	Message   string
	Selection Selection
}

func (s *Selection) IsValid() bool {
	return s.Range.IsValid()
}

func (info *Info) Error() string {
	buf := bytes.Buffer{}
	buf.Grow(len(info.Title))

	if strings.TrimSpace(info.Title) == "" {
		buf.WriteString(emptyMessage)
	} else {
		buf.WriteString(info.Title)
	}

	if info.Selection.Hint != "" {
		buf.WriteString(" (")
		buf.WriteString(info.Selection.Hint)
		buf.WriteString(")")
	}

	return buf.String()
}

func (info *Info) Report() {
	if info.Level > MinDisplayLevel {
		return
	}

	buf := bytes.Buffer{}
	buf.Grow(initialReportBufferSize)

	writeLabel(&buf, info.Level, info.Tag)

	if info.Title != "" {
		titleStyle.Fprint(&buf, info.Title, "\n")
	} else {
		titleStyle.Fprint(&buf, emptyMessage, "\n")
	}

	writeSelection(&buf, info.Selection, levelColor(info.Level))

	for _, suggestion := range info.Suggestions {
		writeSuggestion(&buf, suggestion)
	}

	Output.Write(buf.Bytes())
}

func (info *Info) Info() *Info {
	return info
}

func (info *Info) With(suggestion Suggestion) *Info {
	info.Suggestions = append(info.Suggestions, suggestion)
	return info
}

func writeLabel(buf *bytes.Buffer, level Level, tag string) {
	color := levelColor(level)

	if tag != "" {
		color.Fprintf(buf, "%s %s: ", tag, level.String())
	} else {
		color.Fprintf(buf, "%s: ", level.String())
	}
}

func levelColor(level Level) *color.Color {
	switch level {
	case LevelDebug:
		return color.New(color.Bold, color.FgHiMagenta)

	case LevelHint:
		return color.New(color.Bold, color.FgHiGreen)

	case LevelWarning:
		return color.New(color.Bold, color.FgHiYellow)

	case LevelError:
		return color.New(color.Bold, color.FgHiRed)

	default:
		panic("invalid enum value")
	}
}

func writeSelection(buf *bytes.Buffer, selection Selection, color *color.Color) {
	// TODO handle custom content
	file := config.File(selection.Range.ID())

	if file == nil {
		if selection.Hint != "" {
			buf.WriteString(" (")
			buf.WriteString(selection.Hint)
			buf.WriteString(")\n")
		}

		return
	}

	if position, valid := file.GetPosition(selection.Range.From); valid {
		buf.WriteByte('\n')
		writePosition(buf, position)
		buf.WriteByte('\n')

		if ShowCodeSnapshot {
			endPosition, valid := file.GetPosition(selection.Range.To)

			if !valid {
				panic("selection range is corrupted, invalid right bound position")
			}

			code := file.Line(position.Pos)
			writeCodeSnapshot(buf, color, code, selection.Hint, position, endPosition)
		}
	}
}

func writeCodeSnapshot(buf *bytes.Buffer, color *color.Color, code, hint string, start, end text.Position) {
	left, right := start.Char-1, end.Char-1

	if end.Line > start.Line {
		right = len(code)
	}

	writeLineNumber(buf, start.Line, false)
	writeColoredRange(buf, code, color, codeStyle, left, right)
	buf.WriteByte('\n')

	writeLineNumber(buf, start.Line, true)

	// Underline
	{
		// keep tabs in the output
		for _, c := range code[:left] {
			if c == '\t' {
				buf.WriteRune('\t')
			} else {
				buf.WriteByte(' ')
			}
		}

		underlineLen := max(1, right-left+1)

		buf.WriteByte('^')
		for range underlineLen - 1 {
			buf.WriteByte('~')
		}
	}

	if hint != "" {
		buf.WriteByte(' ')
		buf.WriteString(hint)
	}

	buf.WriteByte('\n')
}

func writeSuggestion(buf *bytes.Buffer, suggestion Suggestion) {
	if suggestion.Message == "" {
		return
	}

	suggestionStyle.Fprint(buf, suggestion.Message, "\n")

	// TODO content must be indented
	const hintIndent = "\t"
	// hintIndent + strings.ReplaceAll(hint.Suggestion, "\n", "\n"+hintIndent)
	writeSelection(buf, suggestion.Selection, suggestionStyle)
}

func writeLineNumber(buf *bytes.Buffer, line int, empty bool) {
	if empty {
		for range numLen(line) {
			buf.WriteByte(' ')
		}
	} else {
		lineNumStyle.Fprint(buf, line)
	}

	if UseUnicode {
		separatorStyle.Fprint(buf, " │")
	} else {
		separatorStyle.Fprint(buf, " |")
	}
}

func writeColoredRange(buf *bytes.Buffer, text string, selection, rest *color.Color, i, j int) {
	if len(text) == 0 {
		return
	}

	n := len(text) - 1
	i = min(n, max(i, 0))
	j = min(n, j)

	if i > j {
		i, j = j, i
	}

	textBefore := text[:i]
	textAfter := text[j+1:]

	rest.Fprint(buf, textBefore)
	selection.Fprint(buf, text[i:j+1])
	rest.Fprint(buf, textAfter)
}

func writePosition(buf *bytes.Buffer, position text.Position) {
	for range numLen(position.Line) {
		buf.WriteByte(' ')
	}

	lineNumStyle.Fprintf(buf, "")

	if UseUnicode {
		separatorStyle.Fprint(buf, " ┌─ ")
	} else {
		separatorStyle.Fprint(buf, "--> ")
	}

	// if !ShowCodeSnapshot {
	// 	buf.WriteString(" ↪ ")
	// }

	path := filepath.Clean(position.Path)

	switch LineInfoStyle {
	default: // LineInfoUnix
		filepathStyle.Fprintf(buf, "%s:%d:%d", path, position.Line, position.Char)
	}
}

func numLen(num int) (len int) {
	if num <= 0 {
		len = 1
	}
	for num != 0 {
		num /= 10
		len += 1
	}
	return len
}

var (
	titleStyle      = color.New(color.FgWhite)
	lineNumStyle    = color.New(color.FgHiCyan, color.Bold)
	separatorStyle  = color.New(color.FgBlack, color.Bold)
	filepathStyle   = color.New(color.FgCyan)
	codeStyle       = color.New(color.FgHiMagenta)
	suggestionStyle = color.New(color.FgWhite)
)
