package report

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
)

func ConfigLevel(cfg *config.Config) Level {
	switch {
	case cfg.Flags.Debug:
		return LevelDebug

	case cfg.Flags.NoHints:
		return LevelWarning

	default:
		return LevelHint
	}
}

//go:generate stringer -type=Level -linecomment
type Level byte

const (
	LevelError   Level = iota // error
	LevelWarning              // warning
	LevelHint                 // hint
	LevelDebug                // debug
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

var colors = [...]*color.Color{
	LevelDebug:   color.New(color.Bold, color.FgHiMagenta),
	LevelHint:    color.New(color.Bold, color.FgHiGreen),
	LevelWarning: color.New(color.Bold, color.FgHiYellow),
	LevelError:   color.New(color.Bold, color.FgHiRed),
}
