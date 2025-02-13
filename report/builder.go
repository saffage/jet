package report

import (
	"fmt"

	"github.com/saffage/jet/text"
)

type builder struct {
	inner error
	info  *Info
}

func Build(inner error) builder {
	return builder{
		inner: inner,
		info:  &Info{Title: inner.Error()},
	}
}

func (b builder) Level(level Level) builder {
	b.info.Level = level
	return b
}

func (b builder) Tag(tag string) builder {
	b.info.Tag = tag
	return b
}

func (b builder) CustomSelection(span text.Span, content, hint string) builder {
	b.info.Selection = Selection{
		CustomContent: content,
		Hint:          hint,
		Range:         span,
	}
	return b
}

func (b builder) CustomSelectionF(span text.Span, content, format string, args ...any) builder {
	b.info.Selection = Selection{
		CustomContent: content,
		Hint:          fmt.Sprintf(format, args...),
		Range:         span,
	}
	return b
}

func (b builder) Selection(span text.Span, hint string) builder {
	b.info.Selection = Selection{
		Hint:  hint,
		Range: span,
	}
	return b
}

func (b builder) SelectionF(span text.Span, format string, args ...any) builder {
	b.info.Selection = Selection{
		Hint:  fmt.Sprintf(format, args...),
		Range: span,
	}
	return b
}

func (b builder) SelectionV(selection Selection) builder {
	b.info.Selection = selection
	return b
}

func (b builder) Suggestion(span text.Span, message string) builder {
	b.info.Suggestions = append(b.info.Suggestions, Suggestion{
		Message:   message,
		Selection: Selection{Range: span},
	})
	return b
}

func (b builder) SuggestionF(span text.Span, format string, args ...any) builder {
	b.info.Suggestions = append(b.info.Suggestions, Suggestion{
		Message:   fmt.Sprintf(format, args...),
		Selection: Selection{Range: span},
	})
	return b
}

func (b builder) SuggestionV(suggestion Suggestion) builder {
	b.info.Suggestions = append(b.info.Suggestions, suggestion)
	return b
}

func (b builder) Error() string { return b.info.Title }

func (b builder) Info() *Info { return b.info }

func (b builder) Unwrap() error { return b.inner }
