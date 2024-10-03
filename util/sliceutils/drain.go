package sliceutils

import "iter"

func Drain[T any](s0 *[]T) iter.Seq[T] {
	return func(yield func(T) bool) {
		s := (*s0)
		*s0 = nil
		for i := range s {
			if !yield(s[i]) {
				break
			}
		}
	}
}
