package report

import (
	"fmt"

	"github.com/saffage/jet/text"
)

type Builder struct {
	inner error
	info  *Info
}

func Build(inner error) Builder {
	return Builder{
		inner: inner,
		info:  &Info{Title: inner.Error()},
	}
}

func (b Builder) Level(level Level) Builder {
	b.info.Level = level
	return b
}

func (b Builder) Tag(tag string) Builder {
	b.info.Tag = tag
	return b
}

func (b Builder) CustomSelection(span text.Span, content, hint string) Builder {
	b.info.Selection = Selection{
		CustomContent: content,
		Hint:          hint,
		Range:         span,
	}
	return b
}

func (b Builder) CustomSelectionF(span text.Span, content, format string, args ...any) Builder {
	b.info.Selection = Selection{
		CustomContent: content,
		Hint:          fmt.Sprintf(format, args...),
		Range:         span,
	}
	return b
}

func (b Builder) Selection(span text.Span, hint string) Builder {
	b.info.Selection = Selection{
		Hint:  hint,
		Range: span,
	}
	return b
}

func (b Builder) SelectionF(span text.Span, format string, args ...any) Builder {
	b.info.Selection = Selection{
		Hint:  fmt.Sprintf(format, args...),
		Range: span,
	}
	return b
}

func (b Builder) SelectionV(selection Selection) Builder {
	b.info.Selection = selection
	return b
}

func (b Builder) Suggestion(message string) Builder {
	b.info.Suggestions = append(b.info.Suggestions, Suggestion{
		Message: message,
	})
	return b
}

func (b Builder) SuggestionF(format string, args ...any) Builder {
	b.info.Suggestions = append(b.info.Suggestions, Suggestion{
		Message: fmt.Sprintf(format, args...),
	})
	return b
}

func (b Builder) SuggestionV(suggestion Suggestion) Builder {
	b.info.Suggestions = append(b.info.Suggestions, suggestion)
	return b
}

func (b Builder) Error() string { return b.info.Title }

func (b Builder) Info() *Info { return b.info }

func (b Builder) Unwrap() error { return b.inner }
