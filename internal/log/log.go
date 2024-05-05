package log

import (
	"fmt"
	"os"
)

func Note(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindNote.Label()}, args...)...)
}

func Hint(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindHint.Label()}, args...)...)
}

func Warn(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindWarning.Label()}, args...)...)
}

func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", append([]any{KindError.Label()}, args...)...)
}

func InternalError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", append([]any{KindInternalError.Label()}, args...)...)
}
