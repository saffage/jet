package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/util"
)

type Info struct {
	Kind         ProblemKind
	Tag          string
	Title        string
	Hint         string
	Descriptions []Description
	Ast          ast.Node
	Range        token.Range // When .IsValid() returns true, it will produce code snapshot
}

// Note is used in [Info] for a more flexible output.
type Description struct {
	Message string
	Ast     ast.Node // unused
	Range   token.Range
}

//go:generate stringer -type=ProblemKind -linecomment -output=problem_kind_string.go

type ProblemKind byte

const (
	KindError   ProblemKind = iota // error
	KindWarning                    // warning
)

func (info *Info) Error() string {
	buf := strings.Builder{}

	if info.Tag != "" {
		buf.WriteString(info.Tag)
		buf.WriteByte(' ')
		buf.WriteString(info.Kind.String())
		buf.WriteString(": ")
	}

	if info.Title != "" {
		buf.WriteString(info.Title)
	} else {
		buf.WriteString("unknown problem")
	}

	if info.Hint != "" {
		buf.WriteString(" (")
		buf.WriteString(info.Hint)
		buf.WriteString(")")
	}

	return buf.String()
}

func (info *Info) Report() {
	switch info.Kind {
	case KindError:
		info.reportWithCodeSnapshot(LevelError)

	case KindWarning:
		info.reportWithCodeSnapshot(LevelWarning)

	default:
		panic("unreachable")
	}
}

func report(level Level, tag, message string) {
	if level < MinDisplayLevel {
		return
	}

	if strings.TrimSpace(message) == "" {
		message = "(no message provided)"
	}

	if UseColors {
		message = messageStyle.Sprint(message)
	}

	fmt.Fprintln(os.Stderr, level.Label(tag), message)
}

func (info *Info) reportWithCodeSnapshot(level Level) {
	if level < MinDisplayLevel {
		return
	}

	// We do it here because the message will not be empty if range is specified.
	if strings.TrimSpace(info.Title) == "" {
		info.Title = "(no message provided)"
	}

	buf := strings.Builder{}
	buf.WriteByte('\n')

	fileInfo, ok := config.Global.Files[info.Range.FileID]

	if ok {
		buf.WriteString(genCodeSnapshot(level, info.Hint, info.Range, fileInfo))
	}

	buf.WriteString(genDescription(info.Descriptions, fileInfo))

	if UseColors {
		report(level, info.Tag, info.Title+buf.String())
	} else {
		report(level, info.Tag, info.Title+buf.String())
	}
}

func genDescription(
	descriptions []Description,
	fileInfo config.FileInfo,
) string {
	buf := strings.Builder{}

	for _, desc := range descriptions {
		buf.WriteByte('\n')
		buf.WriteString(LevelHint.Label(""))
		buf.WriteByte(' ')

		if UseColors {
			buf.WriteString(messageStyle.Sprint(desc.Message))
		} else {
			buf.WriteString(desc.Message)
		}

		if desc.Range.IsValid() {
			codeSnapshot := genCodeSnapshot(LevelHint, "", desc.Range, fileInfo)

			buf.WriteByte('\n')
			buf.WriteString(codeSnapshot)
		}
	}

	return buf.String()
}

func genCodeSnapshot(
	level Level,
	hint string,
	rng token.Range,
	fileInfo config.FileInfo,
) string {
	if !rng.IsValid() || rng.Start.Char == 0 || rng.End.Char == 0 {
		return ""
	}

	var (
		codeSnapshot = fileInfo.Line(int(rng.Start.Line))
		lineNumStr   = fmt.Sprintf("%d", rng.Start.Line)
		emptyLineNum = genLineNum(strings.Repeat(" ", util.NumLen(rng.Start.Line)))
		leftBound    = int(rng.Start.Char) - 1
		rightBound   = int(rng.End.Char) - 1
		buf          = strings.Builder{}
	)

	if rng.End.Line > rng.Start.Line {
		rightBound = len(codeSnapshot)
	}

	buf.WriteString(formatPos(rng.StartPos()))
	buf.WriteByte('\n')

	if ShowCodeSnapshot {
		codeSnapshot := applyColorInRange(
			level.Color(),
			codeSnapshot,
			int(leftBound),
			int(rightBound),
		)

		buf.WriteString(genLineNum(lineNumStr))
		buf.WriteString(codeSnapshot)
		buf.WriteByte('\n')
	} else {
		return buf.String()
	}

	// Tabulation has a variable length, so we need to
	// keep them in a string there.
	underlineLen := max(1, rightBound-leftBound+1)
	underlineStr := "^" + strings.Repeat("~", underlineLen-1)
	underlineBuf := strings.Builder{}
	underlineBuf.Grow(leftBound + underlineLen)

	for _, c := range codeSnapshot[:leftBound] {
		if c == '\t' {
			// TODO fix line shift
			underlineBuf.WriteRune(c)
		} else {
			underlineBuf.WriteByte(' ')
		}
	}

	if UseColors {
		underlineBuf.WriteString(level.Color().Sprint(underlineStr))
	} else {
		underlineBuf.WriteString(underlineStr)
	}

	buf.WriteString(emptyLineNum)
	buf.WriteString(underlineBuf.String())

	if hint != "" {
		buf.WriteByte(' ')

		if UseColors {
			buf.WriteString(level.Color().Sprint(hint))
		} else {
			buf.WriteString(hint)
		}
	}

	buf.WriteByte('\n')
	return buf.String()
}

func genLineNum(text string) string {
	if UseColors {
		return lineNumStyle.Sprintf("%s │", text)
	}
	return text + " │"
}

func applyColorInRange(color *color.Color, text string, i, j int) string {
	if !UseColors {
		return text
	}

	if text == "" {
		return ""
	}

	n := len(text) - 1
	textBefore := text[:max(0, min(i-1, n)+1)]
	textAfter := ""

	if j < n {
		textAfter = text[j+1:]
	}

	return textBefore + color.Sprint(text[i:min(j, n)+1]) + textAfter
}

func formatPos(pos token.Pos) string {
	space := strings.Repeat(" ", util.NumLen(pos.Line))
	s := " ┌─ "

	if !ShowCodeSnapshot {
		s = " ↪ "
	}

	if UseColors {
		return space +
			lineNumStyle.Sprint(s) +
			filepathStyle.Sprint(pos.String())
	}

	return space + s + pos.String()
}

var (
	lineNumStyle  = color.New(color.FgHiCyan, color.Bold)
	filepathStyle = color.New(color.FgCyan)
	messageStyle  = color.New(color.FgWhite, color.Bold)
)
