// This package is only used for quick code prototyping (for the entire project).
package assert

import (
	"fmt"
	"runtime"
	"strings"
)

type Fail struct {
	Msg string
}

func (e Fail) Error() string { return e.Msg }

// Panics with [assert.Fail] if not `ok`.
func Ok(ok bool, message ...any) {
	if !ok {
		b := strings.Builder{}

		if _, file, line, ok := runtime.Caller(1); ok {
			b.WriteString(fmt.Sprintf("%s:%d: ", file, line))
		}

		b.WriteString("assertion failed")

		if m := fmt.Sprint(message...); len(m) > 0 {
			b.WriteString(": ")
			b.WriteString(m)
		}

		panic(Fail{b.String()})
	}
}
