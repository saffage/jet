package token

import (
	"encoding/json"
	"fmt"
)

func (kind Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind *Kind) UnmarshalJSON(data []byte) error {
	name := string(data)
	k := KindFromString(name)

	if k == Illegal {
		return fmt.Errorf("invalid kind: '%s'", name)
	}

	*kind = k
	return nil
}
