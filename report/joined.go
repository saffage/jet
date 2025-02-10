package report

import "strings"

func Join(errs ...error) error {
	notNilErrs := make([]error, 0, len(errs))

	for _, err := range errs {
		if err != nil {
			notNilErrs = append(notNilErrs, err)
		}
	}

	switch len(notNilErrs) {
	case 0:
		return nil

	case 1:
		return notNilErrs[0]

	default:
		return joined(errs)
	}
}

type joined []error

func (err joined) Error() string {
	buf := strings.Builder{}
	buf.WriteString(err[0].Error())

	for _, err := range err {
		buf.WriteByte('\n')
		buf.WriteString(err.Error())
	}

	return buf.String()
}

func (err joined) Unwrap() []error { return err }
func (err joined) Info() *Info     { return nil }
