package report

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/internal/log"
	"github.com/saffage/jet/scanner/base"
	"github.com/saffage/jet/token"
)

func generateReport(kind log.Kind, start, end token.Loc, buffer []byte) string {
	const lineNumSuffix = " |"

	lineNumStr := fmt.Sprintf("%d", start.Line)
	linePrefix := strings.Repeat(" ", max(0, len(lineNumStr)-1))
	buf := strings.Builder{}

	buf.WriteString(formatFilepath(linePrefix, start))

	emptyLineNumStr := strings.Repeat(" ", len(lineNumStr))
	emptyLineNum := emptyLineNumStr + lineNumStyle.Sprint(lineNumSuffix)

	if start.FileID != 0 && start.Line > 0 {
		scanner := base.New(buffer, start.FileID)
		lineContent := scanner.GetLine(start.Line)
		leftBound := start.Char - 1
		rightBound := end.Char - 1

		if end.Line > start.Line {
			rightBound = len(lineContent)
		}

		line := applyColorInRange(
			kind.Style().Add(color.Underline),
			scanner.GetLine(start.Line),
			leftBound,
			rightBound,
		)

		buf.WriteString(lineNumStyle.Sprint(lineNumStr + lineNumSuffix))
		buf.WriteString(line)
		buf.WriteByte('\n')

		underlineStr := string(underlineChar(kind))
		underline := strings.Repeat(underlineStr, max(1, rightBound-leftBound+1))

		buf.WriteString(emptyLineNum)
		buf.WriteString(strings.Repeat(" ", leftBound))
		buf.WriteString(kind.Style().Sprint(underline))
	}

	return buf.String()
}

func applyColorInRange(color *color.Color, text string, a, b int) string {
	if len(text) == 0 {
		return ""
	}

	maxIdx := len(text) - 1
	textBefore, textAfter := text[:max(0, min(a-1, maxIdx)+1)], ""

	if b < maxIdx {
		textAfter = text[b+1:]
	}

	return textBefore + color.Sprint(text[a:min(b, maxIdx)+1]) + textAfter
}

func underlineChar(kind log.Kind) rune {
	switch kind {
	case log.KindNote, log.KindHint, log.KindWarning:
		return '^'

	case log.KindError:
		return '~'

	default:
		panic("unreachable")
	}
}

// Just for clarity.
func formatFilepath(prefix string, loc token.Loc) string {
	return fmt.Sprintf("%s %s %s\n", prefix, lineNumStyle.Sprint("-->"), loc.String())
}

var (
	messageStyle = color.New(color.Bold)
	lineNumStyle = color.New(color.Bold, color.FgHiGreen)
)
