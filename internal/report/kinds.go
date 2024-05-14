package report

import (
	"fmt"

	"github.com/fatih/color"
)

type Kind byte

const (
	KindDebug Kind = iota
	KindNote
	KindHint
	KindWarning
	KindError
)

func (kind Kind) String() string {
	return kinds[kind]
}

func (kind Kind) Color() *color.Color {
	return colors[kind]
}

// Label format is `kind:`.
func (kind Kind) Label() string {
	if UseColors {
		return kind.Color().Sprintf(
			"%"+align()+"s:",
			kind.String(),
		)
	}
	return fmt.Sprintf(
		"%"+align()+"s:",
		kind.String(),
	)
}

// Label format is `kind(tag):` if 'tag' is not empty.
func (kind Kind) TaggedLabel(tag string) string {
	if tag != "" {
		if UseColors {
			return kind.Color().Sprintf("%"+align()+"s(%s):", kind.String(), tag)
		}
		return fmt.Sprintf("%"+align()+"s(%s):", kind.String(), tag)
	}
	return kind.Label()
}

func align() string {
	if alignLabel {
		return fmt.Sprintf("%d", len("warning"))
	}
	return ""
}

var alignLabel = true

var kinds = map[Kind]string{
	KindDebug:   "debug",
	KindNote:    "note",
	KindHint:    "hint",
	KindWarning: "warning",
	KindError:   "error",
}

var colors = map[Kind]*color.Color{
	KindDebug:   color.New(color.Bold, color.FgHiMagenta),
	KindNote:    color.New(color.Bold, color.FgHiBlue),
	KindHint:    color.New(color.Bold, color.FgHiCyan),
	KindWarning: color.New(color.Bold, color.FgHiYellow),
	KindError:   color.New(color.Bold, color.FgHiRed),
}
