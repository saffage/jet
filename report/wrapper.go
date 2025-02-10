package report

import "fmt"

func Wrap(err error, msg string) error {
	if msg == "" {
		return err
	}

	return wrapper{message: msg, inner: err}
}

func Wrapf(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

type wrapper struct {
	inner   error
	message string
}

func (err wrapper) Error() string { return err.message }
func (err wrapper) Unwrap() error { return err.inner }
func (err wrapper) Info() *Info   { return &Info{Title: err.message} }
