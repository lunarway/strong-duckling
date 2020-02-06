package vici

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapConnections(t *testing.T) {
	tt := []struct {
		name   string
		input  map[string]interface{}
		output map[string]IKEConf
	}{
		{
			name: "no local sections",
			input: map[string]interface{}{
				"gw-gw": map[string]interface{}{
					"version":      "IKEv2",
					"local_addrs":  []string{"10.0.224.131", "3.248.95.39"},
					"remote_addrs": []string{"13.74.42.140"},
					"remote-1": map[string]interface{}{
						"id": "13.74.42.140",
					},
				},
			},
			output: map[string]IKEConf{
				"gw-gw": IKEConf{
					IKEVersion:       "IKEv2",
					LocalAddresses:   []string{"10.0.224.131", "3.248.95.39"},
					RemoteAddresses:  []string{"13.74.42.140"},
					LocalAuthSection: nil,
					RemoteAuthSection: map[string]AuthConf{
						"remote-1": AuthConf{
							IKEIdentity: "13.74.42.140",
						},
					},
				},
			},
		},
		{
			name: "no sections",
			input: map[string]interface{}{
				"gw-gw": map[string]interface{}{
					"version":      "IKEv2",
					"local_addrs":  []string{"10.0.224.131", "3.248.95.39"},
					"remote_addrs": []string{"13.74.42.140"},
				},
			},
			output: map[string]IKEConf{
				"gw-gw": IKEConf{
					IKEVersion:        "IKEv2",
					LocalAddresses:    []string{"10.0.224.131", "3.248.95.39"},
					RemoteAddresses:   []string{"13.74.42.140"},
					LocalAuthSection:  nil,
					RemoteAuthSection: nil,
				},
			},
		},
		{
			name: "local and remote sections",
			input: map[string]interface{}{
				"gw-gw": map[string]interface{}{
					"version":      "IKEv2",
					"local_addrs":  []string{"10.0.224.131", "3.248.95.39"},
					"remote_addrs": []string{"13.74.42.140"},
					"local-1": map[string]interface{}{
						"id": "3.248.95.39",
					},
					"remote-1": map[string]interface{}{
						"id": "13.74.42.140",
					},
				},
			},
			output: map[string]IKEConf{
				"gw-gw": IKEConf{
					IKEVersion:      "IKEv2",
					LocalAddresses:  []string{"10.0.224.131", "3.248.95.39"},
					RemoteAddresses: []string{"13.74.42.140"},
					LocalAuthSection: map[string]AuthConf{
						"local-1": AuthConf{
							IKEIdentity: "3.248.95.39",
						},
					},
					RemoteAuthSection: map[string]AuthConf{
						"remote-1": AuthConf{
							IKEIdentity: "13.74.42.140",
						},
					},
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := mapConnections(tc.input)
			if !assert.NoError(t, err, "unexpected output error") {
				return
			}
			assert.Equal(t, tc.output, conn, "output not as expected")
		})
	}
}
