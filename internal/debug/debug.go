package debug

import "fmt"

const Enabled = true

func Assert(ok bool, message ...any) {
	if !ok {
		if len(message) > 0 {
			panic("assertion failed, " + fmt.Sprint(message...))
		} else {
			panic("assertion failed")
		}
	}
}
