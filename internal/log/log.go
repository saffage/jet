package log

import (
	"fmt"
	"os"
)

func Note(format string, message ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindNote.Label()}, message...)...)
}

func Hint(format string, message ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindHint.Label()}, message...)...)
}

func Warn(format string, message ...any) {
	fmt.Fprintf(os.Stdout, "%s "+format+"\n", append([]any{KindWarning.Label()}, message...)...)
}

func Error(format string, message ...any) {
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", append([]any{KindError.Label()}, message...)...)
}

func InternalError(format string, message ...any) {
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", append([]any{KindInternalError.Label()}, message...)...)
}
