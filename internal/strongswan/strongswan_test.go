package strongswan

import (
	"testing"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCollectSasStats(t *testing.T) {
	tt := []struct {
		name              string
		connectionConfigs map[string]vici.IKEConf
		sas               map[string]vici.IkeSa
		expected          []IKESAStatus
	}{
		{
			name: "connection missing from config",
			connectionConfigs: map[string]vici.IKEConf{
				"gw-gw": {
					Unique: "id-1",
					Children: map[string]vici.ChildSAConf{
						"net-net-0": {},
					},
				},
			},
			expected: []IKESAStatus{
				{
					Name: "gw-gw",
					Configuration: vici.IKEConf{
						Unique: "id-1",
						Children: map[string]vici.ChildSAConf{
							"net-net-0": {},
						},
					},
					ChildSA: []ChildSAStatus{
						{Name: "net-net-0"},
					},
				},
			},
		},
		{
			name: "maps correctly",
			connectionConfigs: map[string]vici.IKEConf{
				"gw-gw": {
					Unique: "id-1",
					Children: map[string]vici.ChildSAConf{
						"net-net-0": {},
					},
				},
			},
			sas: map[string]vici.IkeSa{
				"gw-gw": {
					ChildSAs: map[string]vici.ChildSA{
						"net-net-0": {},
					},
				},
			},
			expected: []IKESAStatus{
				{
					Name: "gw-gw",
					Configuration: vici.IKEConf{
						Unique: "id-1",
						Children: map[string]vici.ChildSAConf{
							"net-net-0": {},
						},
					},
					State: &vici.IkeSa{
						ChildSAs: map[string]vici.ChildSA{
							"net-net-0": {},
						},
					},
					ChildSA: []ChildSAStatus{
						{
							Name:  "net-net-0",
							State: &vici.ChildSA{},
						},
					},
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ikeSAStatusReceiver := MockIKESAStatusReceiver{}
			ikeSAStatusReceiver.Test(t)
			var actualStatuses []IKESAStatus
			ikeSAStatusReceiver.On("IKESAStatus", mock.Anything).Run(func(args mock.Arguments) {
				status := args[0].(IKESAStatus)
				actualStatuses = append(actualStatuses, status)
			})

			collectSasStats(tc.connectionConfigs, tc.sas, []IKESAStatusReceiver{&ikeSAStatusReceiver})

			assert.Equal(t, tc.expected, actualStatuses, "IKESAStatuses not as expected")
		})
	}
}
