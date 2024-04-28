package constant

type Value interface {
	Kind() Kind

	String() string

	implValue()
}

func (boolValue) implValue()   {}
func (intValue) implValue()    {}
func (floatValue) implValue()  {}
func (stringValue) implValue() {}
func (*expression) implValue() {}
