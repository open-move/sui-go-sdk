package graphql

import (
	"encoding/json"
)

// UnmarshalJSON handles union types for TransactionInput.
func (t *TransactionInput) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["__typename"]; ok {
		json.Unmarshal(v, &t.Typename)
	}

	switch t.Typename {
	case "SharedInput":
		t.SharedInput = &SharedInput{}
		return json.Unmarshal(data, t.SharedInput)
	case "MoveValue":
		t.MoveValue = &MoveValue{}
		return json.Unmarshal(data, t.MoveValue)
	case "PureInput":
		t.Pure = &PureInput{}
		return json.Unmarshal(data, t.Pure)
	case "ObjectInput":
		t.Object = &ObjectInput{}
		return json.Unmarshal(data, t.Object)
	}
	return nil
}
