package constant

func NewString(value string) Value {
	return stringValue{value}
}

type stringValue struct {
	value string
}

func (v stringValue) Kind() Kind     { return String }
func (v stringValue) String() string { return "\"" + v.value + "\"" }
