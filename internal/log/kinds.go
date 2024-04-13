package log

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var NoColors = false

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

// Label format is "kind:".
func (kind Kind) Label() string {
	return kind.LabelWithPayload("")
}

// Label format is "kind[payload]:" if `payload` is not empty.
func (kind Kind) LabelWithPayload(payload string) string {
	label := strings.Builder{}
	label.WriteString(kind.String())

	if payload != "" {
		label.WriteString(fmt.Sprintf("[%s]", payload))
	}

	label.WriteByte(':')

	if NoColors {
		return label.String()
	}

	return kind.Style().Sprint(label.String())
}

func (kind Kind) Style() *color.Color {
	switch kind {
	case KindNote:
		return color.New(color.Bold, color.FgHiBlue)

	case KindHint:
		return color.New(color.Bold, color.FgCyan)

	case KindWarning:
		return color.New(color.Bold, color.FgYellow)

	case KindError, KindInternalError:
		return color.New(color.Bold, color.FgRed)

	default:
		panic("unreachable")
	}
}

var kinds = map[Kind]string{
	KindNote:          "note",
	KindHint:          "hint",
	KindWarning:       "warning",
	KindError:         "error",
	KindInternalError: "internal error",
}
