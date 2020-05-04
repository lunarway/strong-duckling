package vici

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToGeneral(t *testing.T) {
	tt := []struct {
		name     string
		concrete interface{}
		output   map[string]interface{}
	}{
		{
			name: "list sas payload",
			concrete: map[string]interface{}{
				"ike": "ike_name",
			},
			output: map[string]interface{}{
				"ike": "ike_name",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var general map[string]interface{}
			err := convertToGeneral(tc.concrete, &general)
			assert.NoError(t, err, "conversion error")
			assert.Equal(t, tc.output, general)
		})
	}
}
