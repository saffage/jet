package report

import (
	"fmt"

	"github.com/fatih/color"
)

type Level byte

//go:generate stringer -type=Level -linecomment

const (
	LevelDebug   Level = iota // debug
	LevelHint                 // hint
	LevelWarning              // warning
	LevelError                // error
)

// The longest label name without tag (len("warning"))
const align = "%s" // "%7s"

func (l Level) Label(tag string) string {
	if tag != "" {
		if UseColors {
			return l.Color().Sprintf(align+"(%s):", l.String(), tag)
		}
		return fmt.Sprintf(align+"(%s):", l.String(), tag)
	}
	if UseColors {
		return l.Color().Sprintf(align+":", l.String())
	}
	return fmt.Sprintf(align+":", l.String())
}

func (l Level) Color() *color.Color {
	return colors[l]
}

var colors = map[Level]*color.Color{
	LevelDebug:   color.New(color.Bold, color.FgHiMagenta),
	LevelHint:    color.New(color.Bold, color.FgHiGreen),
	LevelWarning: color.New(color.Bold, color.FgHiYellow),
	LevelError:   color.New(color.Bold, color.FgHiRed),
}
