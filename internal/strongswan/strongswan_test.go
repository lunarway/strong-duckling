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
		conf              *vici.IKEConf
		sa                *vici.IkeSa
	}{
		{
			name: "connection missing from config",
			connectionConfigs: map[string]vici.IKEConf{
				"gw-gw": {
					Unique: "id-1",
				},
			},
			conf: &vici.IKEConf{
				Unique: "id-1",
			},
			sa: nil,
		},
		{
			name: "connection missing from config",
			connectionConfigs: map[string]vici.IKEConf{
				"gw-gw": {
					Unique: "id-1",
				},
			},
			conf: &vici.IKEConf{
				Unique: "id-1",
			},
			sa: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ikeSAStatusReceiver := MockIKESAStatusReceiver{}
			ikeSAStatusReceiver.Test(t)
			var reportedConf *vici.IKEConf
			var reportedSA *vici.IkeSa
			ikeSAStatusReceiver.On("IKESAStatus", mock.Anything).Run(func(args mock.Arguments) {
				status := args[0].(IKESAStatus)
				reportedConf = &status.Configuration
				reportedSA = status.State
			})

			collectSasStats(tc.connectionConfigs, tc.sas, []IKESAStatusReceiver{&ikeSAStatusReceiver})

			assert.Equal(t, tc.conf, reportedConf, "IKE Conf not as expected")
			assert.Equal(t, tc.sa, reportedSA, "IKE SA not as expected")
		})
	}
}
