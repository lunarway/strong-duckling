package vici

import (
	"encoding/json"
)

// convertToGeneral converts a concrete type instance to a untype general type.
func convertToGeneral(concrete interface{}, general map[string]interface{}) error {
	b, err := json.Marshal(concrete)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &general)
}

// convertFromGeneral converts a general type instance to a concrete type.
func convertFromGeneral(general interface{}, concrete interface{}) error {
	b, err := json.Marshal(general)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, concrete)
}
