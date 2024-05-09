package report

import (
	"fmt"

	"github.com/fatih/color"
)

type Kind byte

const (
	KindNote Kind = iota
	KindHint
	KindWarning
	KindError
	KindInternalError
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
		return kind.Color().Sprintf(kind.String() + ":")
	}
	return kind.String() + ":"
}

// Label format is `kind(tag):` if 'tag' is not empty.
func (kind Kind) TaggedLabel(tag string) string {
	if tag != "" {
		if UseColors {
			return kind.Color().Sprintf("%s(%s):", kind.String(), tag)
		}
		return fmt.Sprintf("%s(%s):", kind.String(), tag)
	}
	return kind.Label()
}

var kinds = map[Kind]string{
	KindNote:          "note",
	KindHint:          "hint",
	KindWarning:       "warning",
	KindError:         "error",
	KindInternalError: "internal error",
}

var colors = map[Kind]*color.Color{
	KindNote:          color.New(color.Bold, color.FgHiBlue),
	KindHint:          color.New(color.Bold, color.FgCyan),
	KindWarning:       color.New(color.Bold, color.FgYellow),
	KindError:         color.New(color.Bold, color.FgRed),
	KindInternalError: color.New(color.Bold, color.FgRed),
}
