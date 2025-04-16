//go:build no_assertions

package debug

const Enabled = false

func Assert(ok bool, message ...any) {}
