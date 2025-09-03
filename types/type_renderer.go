package types

import (
	"strconv"

	"github.com/saffage/jet/text"
)

func (t *Named) String() string {
	return Render(t)
}

func (t *Alias) String() string {
	return Render(t)
}

func (t *Fn) String() string {
	return Render(t)
}

func (v *TypeVariable) String() string {
	if v.IsMeta() {
		return "'" + strconv.FormatUint(uint64(v.id), 10)
	}
	return "'" + v.name
}

// func (t Descriptor) String() string {
// 	return Render(t)
// }

func (m *Module) String() string {
	return "module"
}

//
//
//

func (t *Named) Render(buf text.Writer) (rendered bool) {
	if t == nil {
		return false
	}

	if name := t.Name(); name == "" {
		buf.WriteString(`"Invalid Type"`)
	} else {
		buf.WriteString(name)
	}

	if len(t.args) > 0 {
		buf.WriteString("(")

		t.args[0].Render(buf)

		for _, arg := range t.args[1:] {
			buf.WriteString(", ")
			arg.Render(buf)
		}

		buf.WriteString(")")
	}

	return true
}

func (t *Alias) Render(buf text.Writer) (rendered bool) {
	if t == nil {
		return false
	}

	if name := t.Name(); name == "" {
		buf.WriteString(`"Invalid Alias"`)
	} else {
		buf.WriteString(name)
	}

	return true
}

func (t *Fn) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("fn(")

	if t.params != nil {
		t.params.Render(buf)
	}

	buf.WriteString(")")

	if t.result != nil {
		buf.WriteByte(' ')

		t.result.Render(buf)
	}

	return true
}

func (t *TypeVariable) Render(buf text.Writer) (rendered bool) {
	buf.WriteByte('\'')

	if t.IsMeta() {
		buf.WriteString(strconv.FormatUint(uint64(t.id), 10))
	} else {
		buf.WriteString(t.name)
	}

	return true
}

// func (t Descriptor) Render(buf text.Writer) (rendered bool) {
// 	buf.WriteString("type ")

// 	t.base.Render(buf)

// 	return true
// }

func (m *Module) Render(buf text.Writer) (rendered bool) {
	buf.WriteString("module")
	return true
}

//
//
//

func (t *Named) Renderable() bool {
	return true
}

func (t *Alias) Renderable() bool {
	return true
}

func (t *Fn) Renderable() bool {
	return true
}

func (t *TypeVariable) Renderable() bool {
	return true
}

// func (t Descriptor) Renderable() bool {
// 	return true
// }

func (m *Module) Renderable() bool {
	return true
}
