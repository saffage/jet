package token

import (
	"encoding/json"
	"fmt"
)

func (kind *Kind) UnmarshalJSON(data []byte) error {
	name := string(data)
	for k, v := range kindNames {
		if v == name {
			*kind = k
			return nil
		}
	}
	return fmt.Errorf("invalid kind: '%s'", name)
}

func (kind Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.Name())
}
