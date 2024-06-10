package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/scanner/base"
	"github.com/saffage/jet/token"
)

func display(kind Kind, message string) {
	err := error(nil)

	switch kind {
	case KindNote, KindHint, KindWarning:
		_, err = fmt.Fprintln(os.Stdout, message)

	case KindDebug, KindError:
		_, err = fmt.Fprintln(os.Stderr, message)

	default:
		panic("unreachable")
	}

	if err != nil {
		panic(err)
	}
}

func reportInternal(kind Kind, tag, message string) {
	if kind == KindDebug && !IsDebug {
		return
	}

	if strings.TrimSpace(message) == "" {
		message = "<no message provided>"
	}

	display(kind, fmt.Sprintf("%s %s", kind.TaggedLabel(tag), message))
}

func reportAtInternal(kind Kind, tag string, start, end token.Loc, message string) {
	if kind == KindDebug && !IsDebug {
		return
	}

	if start.FileID != end.FileID {
		panic(fmt.Sprintf("start & end position have different file IDs (%d and %d)", start.FileID, end.FileID))
	}

	// We do it here because the message will not be empty
	// if the location is specified.
	if strings.TrimSpace(message) == "" {
		message = "<no message provided>"
	}

	line := "\n" + formatLoc(start)

	if config.Global != nil {
		if fileInfo, ok := config.Global.Files[start.FileID]; ok {
			line += generateLine(kind, start, end, fileInfo.Buf.Bytes())
		}
	}

	if UseColors {
		reportInternal(kind, tag, message+line)
	} else {
		reportInternal(kind, tag, message+line)
	}
}

func generateLine(kind Kind, start, end token.Loc, buffer []byte) string {
	if !ShowLine || start.FileID == 0 || start.Line == 0 {
		return ""
	}
	var (
		lineContent  = base.New(buffer, start.FileID).GetLine(int(start.Line))
		lineNumStr   = fmt.Sprintf("%d", start.Line)
		emptyLineNum = lineNum(strings.Repeat(" ", numLen(int(start.Line))))
		leftBound    = int(start.Char) - 1
		rightBound   = int(end.Char) - 1
		buf          = strings.Builder{}
	)
	if end.Line > start.Line {
		// TODO capture more lines?
		rightBound = len(lineContent)
	}

	kindColor := *kind.Color()
	kindColor.Add(color.Underline)

	buf.WriteByte('\n')
	buf.WriteString(lineNum(lineNumStr))
	buf.WriteString(applyColorInRange(
		&kindColor,
		lineContent,
		int(leftBound),
		int(rightBound),
	))
	buf.WriteByte('\n')

	// Tabulation has a variable length, so you need to
	// keep them in a string there.
	underlineLen := max(1, rightBound-leftBound+1)
	underlineLine := strings.Builder{}
	underlineLine.Grow(leftBound + underlineLen)
	for _, c := range lineContent[:leftBound] {
		if c == '\t' {
			underlineLine.WriteRune(c)
		} else {
			underlineLine.WriteByte(' ')
		}
	}

	if UseColors {
		underlineLine.WriteString(kind.Color().Sprintf(
			strings.Repeat(string(underlineChar(kind)), underlineLen),
		))
	} else {
		underlineLine.WriteString(
			strings.Repeat(string(underlineChar(kind)), underlineLen),
		)
	}

	buf.WriteString(emptyLineNum)
	buf.WriteString(underlineLine.String())
	return buf.String()
}

func lineNum(text string) string {
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
	return textBefore + color.Sprint(text[a:min(b, maxIdx)+1]) + textAfter
}

func underlineChar(kind Kind) rune {
	switch kind {
	case KindDebug:
		return '-'

	case KindNote, KindHint, KindWarning:
		return '^'

	case KindError:
		return '~'

	default:
		panic("unreachable")
	}
}

func formatLoc(loc token.Loc) string {
	if UseColors {
		return fmt.Sprintf("%s%s %s",
			strings.Repeat(" ", numLen(int(loc.Line))),
			lineNumStyle.Sprint("-->"),
			color.CyanString(loc.String()),
		)
	}
	return fmt.Sprintf("%s--> %s",
		strings.Repeat(" ", numLen(int(loc.Line))),
		color.CyanString(loc.String()),
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

var lineNumStyle = color.New(color.Bold, color.FgHiGreen)
