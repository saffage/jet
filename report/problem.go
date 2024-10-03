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
	Description string
	Ast         ast.Node // unused
	Range       token.Range
}

//go:generate stringer -type=ProblemKind -linecomment -output=problem_kind_string.go

type ProblemKind byte

const (
	KindError   ProblemKind = iota // error
	KindWarning                    // warning
)

func (p *Info) Error() string {
	buf := strings.Builder{}

	if p.Tag != "" {
		buf.WriteString(p.Tag)
		buf.WriteByte(' ')
		buf.WriteString(p.Kind.String())
		buf.WriteString(": ")
	}

	if p.Title != "" {
		buf.WriteString(p.Title)
	} else {
		buf.WriteString("unknown problem")
	}

	if p.Hint != "" {
		buf.WriteString(": ")
		buf.WriteString(p.Hint)
	}

	return buf.String()
}

func (p *Info) Report() {
	switch p.Kind {
	case KindError:
		reportWithCodeSnapshot(LevelError, *p)

	case KindWarning:
		reportWithCodeSnapshot(LevelWarning, *p)

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

	_, err := fmt.Fprintln(os.Stderr, level.Label(tag)+message)
	if err != nil {
		panic(err)
	}
}

func reportWithCodeSnapshot(level Level, info Info) {
	if level < MinDisplayLevel {
		return
	}

	// We do it here because the message will not be empty if range is specified.
	if strings.TrimSpace(info.Title) == "" {
		info.Title = "(no title provided)"
	}

	line := strings.Builder{}
	line.WriteByte('\n')
	line.WriteString(formatPos(info.Range.StartPos()))

	fileInfo, ok := config.Global.Files[info.Range.FileID]

	if ok && ShowCodeSnapshot {
		line.WriteString(genCodeSnapshot(level, info.Hint, info.Range, fileInfo))
		line.WriteByte('\n')
	}

	for _, desc := range info.Descriptions {
		line.WriteString("\n  ")
		line.WriteString(descriptionStyle.Sprint(desc.Description))

		if desc.Range.IsValid() {
			line.WriteByte('\n')
			line.WriteString(genCodeSnapshot(LevelHint, "", desc.Range, fileInfo))
		}
	}

	if UseColors {
		report(level, info.Tag, info.Title+line.String())
	} else {
		report(level, info.Tag, info.Title+line.String())
	}
}

func genCodeSnapshot(
	level Level,
	description string,
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

	buf.WriteByte('\n')
	buf.WriteString(genLineNum(lineNumStr))
	buf.WriteString(applyColorInRange(
		level.Color(),
		codeSnapshot,
		int(leftBound),
		int(rightBound),
	))
	buf.WriteByte('\n')

	// Tabulation has a variable length, so we need to
	// keep them in a string there.
	underlineLen := max(1, rightBound-leftBound+1)
	underlineLine := strings.Builder{}
	underlineLine.Grow(leftBound + underlineLen)
	for _, c := range codeSnapshot[:leftBound] {
		if c == '\t' {
			// TODO fix line shift
			underlineLine.WriteRune(c)
		} else {
			underlineLine.WriteByte(' ')
		}
	}

	underline := "^" + strings.Repeat("~", underlineLen-1)

	if UseColors {
		underlineLine.WriteString(level.Color().Sprint(underline))
	} else {
		underlineLine.WriteString(underline)
	}

	buf.WriteString(emptyLineNum)
	buf.WriteString(underlineLine.String())

	if description != "" {
		buf.WriteByte(' ')

		if UseColors {
			buf.WriteString(level.Color().Sprint(description))
		} else {
			buf.WriteString(description)
		}
	}

	return buf.String()
}

func genLineNum(text string) string {
	if UseColors {
		return lineNumStyle.Sprintf("%s |", text)
	}
	return text + " |"
}

func applyColorInRange(color *color.Color, text string, a, b int) string {
	if !UseColors {
		return text
	}
	if len(text) == 0 {
		return ""
	}
	maxIdx := len(text) - 1
	textBefore, textAfter := text[:max(0, min(a-1, maxIdx)+1)], ""
	if b < maxIdx {
		textAfter = text[b+1:]
	}
	tmp := text[a : min(b, maxIdx)+1]
	return textBefore + color.Sprint(tmp) + textAfter
}

func formatPos(pos token.Pos) string {
	space := strings.Repeat(" ", util.NumLen(pos.Line))

	if UseColors {
		return fmt.Sprintf("%s%s %s",
			space,
			lineNumStyle.Sprint("-->"),
			filepathStyle.Sprint(pos.String()),
		)
	}

	return fmt.Sprintf("%s--> %s", space, pos.String())
}

var (
	lineNumStyle     = color.New(color.FgHiCyan, color.Bold)
	filepathStyle    = color.New(color.FgCyan)
	descriptionStyle = color.New(color.FgWhite, color.Bold)
)
