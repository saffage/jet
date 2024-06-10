package types

import "fmt"

type Named struct {
	name       string
	underlying Type

	methods map[string]Func
}

func NewNamed(name string, underlying Type) *Named {
	return &Named{name, underlying, nil}
}

func (n *Named) Equals(expected Type) bool {
	if expected != nil {
		if named, _ := expected.(*Named); named != nil {
			return n == named && n.name == named.name
		}
	}
	return false
}

func (n *Named) Underlying() Type { return SkipAlias(n.underlying) }

func (n *Named) String() string { return fmt.Sprintf("%s aka %s", n.name, n.underlying.String()) }

func (n *Named) Name() string { return n.name }

func (n *Named) Methods() map[string]Func { return n.methods }
