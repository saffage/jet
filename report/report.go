package report

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/token"
)

func genHint(
	hints []HintInfo,
	span token.Range,
	file *config.File,
	cfg *config.Config,
) string {
	buf := strings.Builder{}

	for _, hint := range hints {
		buf.WriteByte('\n')
		buf.WriteString(hint.Message)

		if hint.HintRange.IsValid() {
			codeSnapshot := genCodeSnapshot(LevelHint, "", span, file, cfg)

			buf.WriteByte('\n')
			buf.WriteString(codeSnapshot)
		} else {
			const hintIndent = "\t"

			if hint.Suggestion != "" {
				buf.WriteString(":\n\n")
				buf.WriteString(hintIndent)
				buf.WriteString(strings.ReplaceAll(hint.Suggestion, "\n", "\n"+hintIndent))
				buf.WriteString("\n")
			}
		}
	}

	return buf.String()
}

func genCodeSnapshot(
	level Level,
	hint string,
	span token.Range,
	file *config.File,
	cfg *config.Config,
) string {
	if !span.IsValid() {
		return ""
	}

	var (
		codeSnapshot = file.Line(int(span.Start.Line))
		lineNumStr   = fmt.Sprintf("%d", span.Start.Line)
		emptyLineNum = genLineNum(strings.Repeat(" ", numLen(span.Start.Line)))
		leftBound    = int(span.Start.Char) - 1
		rightBound   = int(span.End.Char) - 1
		buf          = strings.Builder{}
	)

	if len(codeSnapshot) == 0 {
		// Line info is corrupted
		panic("unreachable")
	}

	if span.End.Line > span.Start.Line {
		rightBound = len(codeSnapshot)
	}

	buf.WriteString(formatPos(span.StartPos()))
	// buf.WriteByte('\n')

	if !ShowCodeSnapshot {
		// Only line info will be shown
		return buf.String()
	}

	buf.WriteByte('\n')
	buf.WriteString(genLineNum(lineNumStr))
	buf.WriteString(applyColorInRange(
		level.Color(),
		codeSnapshot,
		int(leftBound),
		int(rightBound),
	))

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

	buf.WriteByte('\n')
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

	textBefore = codeStyle.Sprint(textBefore)
	textAfter = codeStyle.Sprint(textAfter)
	return textBefore + color.Sprint(text[i:min(j, n)+1]) + textAfter
}

func formatPos(pos token.Pos) string {
	space := strings.Repeat(" ", numLen(pos.Line))
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

func numLen(num uint32) (len int) {
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
	lineNumStyle  = color.New(color.FgHiCyan, color.Bold)
	filepathStyle = color.New(color.FgCyan)
	codeStyle     = color.New(color.FgHiMagenta)
)
