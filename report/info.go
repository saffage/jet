package report

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/internal/debug"
)

type Info struct {
	Title       string
	Tag         string
	Suggestions []Suggestion
	Selection   Selection
	Level       Level
}

type Selection struct {
	Code  string
	Hint  string
	Range text.Span
}

type Suggestion struct {
	Message   string
	Content   string
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

func (info *Info) Render(buf *strings.Builder) {
	const initialReportBufferSize = 512

	if info.Level > MinDisplayLevel {
		return
	}

	buf.Grow(initialReportBufferSize)

	writeLabel(buf, info.Level, info.Tag)

	if info.Title != "" {
		titleStyle.Fprint(buf, capitalize(info.Title), "\n")
	} else {
		titleStyle.Fprint(buf, emptyMessage, "\n")
	}

	writeSelection(buf, info.Selection, levelColor(info.Level))

	for _, suggestion := range info.Suggestions {
		writeSuggestion(buf, suggestion)
	}
}

func (info *Info) Renderable() bool {
	return true
}

func (info *Info) Info() *Info {
	return info
}

func (info *Info) With(suggestion Suggestion) *Info {
	info.Suggestions = append(info.Suggestions, suggestion)
	return info
}

func writeLabel(buf *strings.Builder, level Level, tag string) {
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

func writeSelection(
	buf *strings.Builder,
	selection Selection,
	color *color.Color,
) {
	if !ShowCodeSnapshot || !selection.Range.IsValid() {
		if selection.Hint != "" {
			buf.WriteString(" (")
			buf.WriteString(selection.Hint)
			buf.WriteString(")\n")
		}

		return
	}

	file := config.File(selection.Range.ID())
	code := ""

	if selection.Code != "" {
		if nl := strings.IndexByte(selection.Code, '\n'); nl < 0 {
			code = selection.Code
		} else {
			code = selection.Code[:nl]
		}
	} else if file != nil {
		// TODO: it could be optimized.
		position := file.PositionOf(selection.Range.From)
		debug.Assert(position.IsValid(), fmt.Sprintf("%#v",position))
		code = file.LineContent(position.Line)
	}

	from := file.PositionOf(selection.Range.From)
	to := file.PositionOf(selection.Range.To)

	if !to.IsValid() {
		if from.IsValid() {
			to = from
		} else {
			panic(fmt.Sprintf(
				"selection range is corrupted, invalid selection bounds (%s to %s), file context length is %d",
				selection.Range.From,
				selection.Range.To,
				len(config.File(selection.Range.ID()).Content),
			))
		}
	}

	writeFilepath(buf, from, code != "")

	buf.WriteByte('\n')

	if code != "" {
		writeCodeSnapshot(buf, color, code, selection.Hint, from, to)
	}
}

func writeCodeSnapshot(
	buf *strings.Builder,
	color *color.Color,
	code, hint string,
	start, end text.Position,
) {
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

		color.Fprint(buf, "^", strings.Repeat("~", underlineLen-1))
	}

	if hint != "" {
		buf.WriteByte(' ')
		color.Fprint(buf, capitalize(hint))
	}

	buf.WriteByte('\n')
}

func writeSuggestion(buf *strings.Builder, suggestion Suggestion) {
	if suggestion.Message == "" {
		return
	}

	suggestionStyle.Fprint(buf, capitalize(suggestion.Message), "\n")

	if suggestion.Content != "" {
		content := "\t" + strings.ReplaceAll(suggestion.Content, "\n", "\n\t")
		suggestionContentStyle.Fprintln(buf, content)
	} else {
		writeSelection(buf, suggestion.Selection, suggestionStyle)
	}
}

func writeLineNumber(buf *strings.Builder, line int, empty bool) {
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

func writeColoredRange(
	buf *strings.Builder,
	text string,
	selection, rest *color.Color,
	i, j int,
) {
	if len(text) == 0 || i >= len(text) || j >= len(text) {
		rest.Fprint(buf, text)
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

func writeFilepath(buf *strings.Builder, position text.Position, withSnapshot bool) {
	for range numLen(position.Line) {
		buf.WriteByte(' ')
	}

	if withSnapshot {
		if UseUnicode {
			separatorStyle.Fprint(buf, " ┌─ ")
		} else {
			separatorStyle.Fprint(buf, "--> ")
		}
	} else {
		if UseUnicode {
			separatorStyle.Fprint(buf, " ↪ ")
		} else {
			separatorStyle.Fprint(buf, "--> ")
		}
	}

	path := filepath.Clean(position.Filepath)

	filepathStyle.Fprintf(
		buf,
		"%s:%d:%d",
		path,
		position.Line,
		position.Char,
	)
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

func capitalize(s string) string {
	firstNonSpace := strings.IndexFunc(s, func(r rune) bool { return r != ' ' })

	if firstNonSpace >= 0 {
		firstNonSpace++
		return strings.ToUpper(s[:firstNonSpace]) + s[firstNonSpace:]
	}

	return s
}

var (
	titleStyle             = color.New(color.FgWhite)
	lineNumStyle           = color.New(color.FgHiCyan, color.Bold)
	separatorStyle         = color.New(color.FgBlack, color.Bold)
	filepathStyle          = color.New(color.FgCyan)
	codeStyle              = color.New(color.FgHiMagenta)
	suggestionStyle        = color.New(color.FgWhite)
	suggestionContentStyle = color.New(color.FgWhite, color.Italic)
)
