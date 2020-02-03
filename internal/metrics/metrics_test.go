package metrics

import (
	"strings"
	"testing"

	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/lunarway/strong-duckling/internal/test"
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var _ tcpchecker.Reporter = &PrometheusReporter{}

func TestIKESAStatus_gauges(t *testing.T) {

	tt := []struct {
		name                                     string
		conf                                     vici.IKEConf
		sa                                       *vici.IkeSa
		packetsIn, packetsOut, bytesIn, bytesOut float64
	}{
		{
			name: "single child sa with packets",
			conf: vici.IKEConf{},
			sa: &vici.IkeSa{
				ChildSAs: map[string]vici.ChildSA{
					"net-1": vici.ChildSA{
						PacketsIn:  "1",
						PacketsOut: "2",
						BytesIn:    "3",
						BytesOut:   "4",
					},
				},
			},
			packetsIn:  1,
			packetsOut: 2,
			bytesIn:    3,
			bytesOut:   4,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			logger := test.NewLogger(t)
			p, err := NewPrometheusReporter(reg, logger)
			if !assert.NoError(t, err, "unexpected initialization error") {
				return
			}

			p.IKESAStatus(tc.conf, tc.sa)

			assert.Equal(t, tc.packetsIn, testutil.ToFloat64(p.ikeSA.packetsIn), "packets in not as expected")
			assert.Equal(t, tc.packetsOut, testutil.ToFloat64(p.ikeSA.packetsOut), "packets out not as expected")
			assert.Equal(t, tc.bytesIn, testutil.ToFloat64(p.ikeSA.bytesIn), "bytes in not as expected")
			assert.Equal(t, tc.bytesOut, testutil.ToFloat64(p.ikeSA.bytesOut), "bytes out not as expected")
		})
	}
}

func TestIKESAStatus_establishedSeconds(t *testing.T) {
	float := func(f float64) *float64 {
		return &f
	}
	tt := []struct {
		name              string
		establishedValues []string
		result            *float64
	}{
		{
			name:              "increasing value",
			establishedValues: []string{"1", "2", "3"},
			result:            nil,
		},
		{
			// we are not sure to know if we have reached the max so we should only
			// set the metric if has decreased, ie. connection reset.
			name:              "set previous when new is below",
			establishedValues: []string{"1", "2", "3", "1"},
			result:            float(3),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			logger := test.NewLogger(t)
			p, err := NewPrometheusReporter(reg, logger)
			if !assert.NoError(t, err, "unexpected initialization error") {
				return
			}

			for _, s := range tc.establishedValues {
				p.IKESAStatus(vici.IKEConf{}, &vici.IkeSa{
					EstablishedSeconds: s,
				})
			}

			if tc.result == nil {
				// this validates that no metrics are collected on the registry
				err = testutil.GatherAndCompare(reg, strings.NewReader(``))
				assert.NoError(t, err, "unexpected error from gathering metrics")
				return
			}
			assert.Equal(t, *tc.result, testutil.ToFloat64(p.ikeSA.establishedSeconds), "establishedSeconds not as expected")
		})
	}
}

func TestIKESAStatus_labels(t *testing.T) {
	tt := []struct {
		name   string
		conf   vici.IKEConf
		sa     *vici.IkeSa
		output string
	}{
		{
			name: "complete label set",
			conf: vici.IKEConf{},
			sa: &vici.IkeSa{
				ChildSAs: map[string]vici.ChildSA{
					"net-1": vici.ChildSA{
						State:      "INSTALLED",
						IPsecMode:  "TUNNEL",
						PacketsIn:  "123",
						PacketsOut: "321",
					},
				},
			},
			output: `# HELP strong_duckling_ike_sa_packets_in_total Total number of received packets
# TYPE strong_duckling_ike_sa_packets_in_total gauge
strong_duckling_ike_sa_packets_in_total{child_sa_name="net-1"} 123
# HELP strong_duckling_ike_sa_packets_out_total Total number of transmitted packets
# TYPE strong_duckling_ike_sa_packets_out_total gauge
strong_duckling_ike_sa_packets_out_total{child_sa_name="net-1"} 321
`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			logger := test.NewLogger(t)
			p, err := NewPrometheusReporter(reg, logger)
			if !assert.NoError(t, err, "unexpected initialization error") {
				return
			}
			p.IKESAStatus(tc.conf, tc.sa)
			err = testutil.GatherAndCompare(reg, strings.NewReader(tc.output))
			assert.NoError(t, err, "registered metrics not as expected")
		})
	}
}
