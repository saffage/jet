package types

type NamedInstance struct {
	named *Named
	args  []Type
}

type AliasInstance struct {
	alias *Alias
	args  []Type
}
