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

var _ tcpchecker.Reporter = (&PrometheusReporter{}).TcpChecker()

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

func TestIKESAStatus_installs(t *testing.T) {
	tt := []struct {
		name               string
		installTimeSeconds []string
		installsSet        bool
		installs           float64
	}{
		{
			name:               "single value",
			installTimeSeconds: []string{"1"},
			installsSet:        false,
			installs:           0,
		},
		{
			name:               "max value",
			installTimeSeconds: []string{"1", "2", "3", "1"},
			installsSet:        true,
			installs:           1,
		},
		{
			name:               "multiple max value",
			installTimeSeconds: []string{"1", "2", "3", "1", "2", "1"},
			installsSet:        true,
			installs:           2,
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

			for _, s := range tc.installTimeSeconds {
				p.IKESAStatus(vici.IKEConf{}, &vici.IkeSa{
					ChildSAs: map[string]vici.ChildSA{
						"net-0": vici.ChildSA{
							InstallTimeSeconds: s,
						},
					},
				})
			}
			if !tc.installsSet {
				// this validates that no metrics are collected on the registry
				err = testutil.GatherAndCompare(reg, strings.NewReader(``))
				assert.NoError(t, err, "unexpected error from gathering metrics")
				return
			}
			assert.Equal(t, tc.installs, testutil.ToFloat64(p.ikeSA.installs), "installs not as expected")
		})
	}
}
func TestIKESAStatus_rekeySeconds(t *testing.T) {
	tt := []struct {
		name             string
		connRekeySeconds string
		rekeySeconds     []string
		rekeySet         bool
		histogram        string
	}{
		{
			name:             "single value",
			connRekeySeconds: "10",
			rekeySeconds:     []string{"1"},
			rekeySet:         false,
			histogram:        "",
		},
		{
			name:             "single min value",
			connRekeySeconds: "100",
			rekeySeconds:     []string{"50", "40", "30", "90"},
			rekeySet:         true,
			histogram: `# HELP strong_duckling_ike_sa_rekey_seconds Duration between re-keying
# TYPE strong_duckling_ike_sa_rekey_seconds histogram
strong_duckling_ike_sa_rekey_seconds_bucket{le="10"} 0
strong_duckling_ike_sa_rekey_seconds_bucket{le="30"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="60"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="120"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="300"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="480"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="600"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="+Inf"} 1
strong_duckling_ike_sa_rekey_seconds_sum 30
strong_duckling_ike_sa_rekey_seconds_count 1
`,
		},
		{
			name:             "multiple min values",
			connRekeySeconds: "100",
			rekeySeconds:     []string{"50", "40", "30", "90", "50", "100"},
			rekeySet:         true,
			histogram: `# HELP strong_duckling_ike_sa_rekey_seconds Duration between re-keying
# TYPE strong_duckling_ike_sa_rekey_seconds histogram
strong_duckling_ike_sa_rekey_seconds_bucket{le="10"} 0
strong_duckling_ike_sa_rekey_seconds_bucket{le="30"} 1
strong_duckling_ike_sa_rekey_seconds_bucket{le="60"} 2
strong_duckling_ike_sa_rekey_seconds_bucket{le="120"} 2
strong_duckling_ike_sa_rekey_seconds_bucket{le="300"} 2
strong_duckling_ike_sa_rekey_seconds_bucket{le="480"} 2
strong_duckling_ike_sa_rekey_seconds_bucket{le="600"} 2
strong_duckling_ike_sa_rekey_seconds_bucket{le="+Inf"} 2
strong_duckling_ike_sa_rekey_seconds_sum 80
strong_duckling_ike_sa_rekey_seconds_count 2
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

			for _, s := range tc.rekeySeconds {
				p.IKESAStatus(vici.IKEConf{
					RekeyTimeSeconds: tc.connRekeySeconds,
				}, &vici.IkeSa{
					ChildSAs: map[string]vici.ChildSA{
						"net-0": vici.ChildSA{
							RekeyTimeSeconds: s,
						},
					},
				})
			}
			// if !tc.rekeySet {
			// this validates that no metrics are collected on the registry
			err = testutil.GatherAndCompare(reg, strings.NewReader(tc.histogram))
			assert.NoError(t, err, "unexpected error from gathering metrics")
			// 	return
			// }
			// assert.Equal(t, tc.rekeyDuration, testutil.ToFloat64(p.ikeSA.rekeySeconds), "rekey not as expected")
		})
	}
}

func TestPrometheusReporter_maxValue(t *testing.T) {
	tt := []struct {
		name   string
		values []string
		output float64
		ok     bool
	}{
		{
			name:   "single value",
			values: []string{"1"},
			output: 1,
			ok:     false,
		},
		{
			name:   "increasing values",
			values: []string{"1", "2", "3"},
			output: 3,
			ok:     false,
		},
		{
			name:   "decreasing values",
			values: []string{"3", "2", "1"},
			output: 2,
			ok:     true,
		},
		{
			name:   "values have a max",
			values: []string{"1", "2", "3", "1"},
			output: 3,
			ok:     true,
		},
		{
			name:   "values have a min",
			values: []string{"3", "2", "1", "2"},
			output: 2,
			ok:     false,
		},
		{
			name:   "values are equal",
			values: []string{"1", "1", "1", "1"},
			output: 1,
			ok:     false,
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

			var v float64
			var ok bool
			for _, s := range tc.values {
				v, ok = p.maxValue("test", s)
			}

			assert.Equal(t, tc.ok, ok, "ok indication not as expected")
			assert.Equal(t, tc.output, v, "value not as expected")
		})
	}
}
func TestPrometheusReporter_minValue(t *testing.T) {
	tt := []struct {
		name   string
		values []string
		output float64
		ok     bool
	}{
		{
			name:   "single value",
			values: []string{"1"},
			output: 1,
			ok:     false,
		},
		{
			name:   "increasing values",
			values: []string{"1", "2", "3"},
			output: 2,
			ok:     true,
		},
		{
			name:   "decreasing values",
			values: []string{"3", "2", "1"},
			output: 1,
			ok:     false,
		},
		{
			name:   "values have a max",
			values: []string{"1", "2", "3", "1"},
			output: 1,
			ok:     false,
		},
		{
			name:   "values have a min",
			values: []string{"3", "2", "1", "2"},
			output: 1,
			ok:     true,
		},
		{
			name:   "values are equal",
			values: []string{"1", "1", "1", "1"},
			output: 1,
			ok:     false,
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

			var v float64
			var ok bool
			for _, s := range tc.values {
				v, ok = p.minValue("test", s)
			}

			assert.Equal(t, tc.ok, ok, "ok indication not as expected")
			assert.Equal(t, tc.output, v, "value not as expected")
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
