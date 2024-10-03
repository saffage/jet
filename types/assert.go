package types

import "fmt"

func assert(ok bool, message ...any) {
	if !ok {
		if len(message) > 0 {
			panic("assertion failed: " + fmt.Sprint(message...))
		}
		panic("assertion failed")
	}
}
