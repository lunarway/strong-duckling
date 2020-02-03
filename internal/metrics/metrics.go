package metrics

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Register(serveMux *http.ServeMux) {
	serveMux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(rw, r)
		}),
	))
}

const (
	namespace           = "strong_duckling"
	subSystemTcpChecker = "tcp_checker"
	subSystemIKE        = "ike_sa"
)

// metric label consts
const (
	strongswanVersion     = "strongswan_version"
	strongDucklingVersion = "version"
	childSAName           = "child_sa_name"
)

type Logger interface {
	Errorf(string, ...interface{})
}

type PrometheusReporter struct {
	registry prometheus.Registerer
	logger   Logger

	version    *prometheus.GaugeVec
	tcpChecker *tcpChecker
	ikeSA      ikeSA
}

func (pr *PrometheusReporter) TcpChecker() tcpchecker.Reporter {
	return pr.tcpChecker
}

type tcpChecker struct {
	open             *prometheus.GaugeVec
	connectedTotal   *prometheus.CounterVec
	disconectedTotal *prometheus.CounterVec

	previousOpenState *bool
}

type ikeSA struct {
	establishedSeconds         *prometheus.GaugeVec
	previousEstablishedSeconds float64
	packetsIn                  *prometheus.GaugeVec
	packetsOut                 *prometheus.GaugeVec
	lastPacketInSeconds        *prometheus.HistogramVec
	lastPacketOutSeconds       *prometheus.HistogramVec
	bytesIn                    *prometheus.GaugeVec
	bytesOut                   *prometheus.GaugeVec
	installedSeconds           *prometheus.CounterVec
	rekeySeconds               *prometheus.HistogramVec
	lifeTimeSeconds            *prometheus.HistogramVec
	state                      *prometheus.GaugeVec
	childSAState               *prometheus.GaugeVec
}

func NewPrometheusReporter(reg prometheus.Registerer, logger Logger) (*PrometheusReporter, error) {
	r := PrometheusReporter{
		registry: reg,
		logger:   logger,
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Version info of strong_duckling",
		}, []string{strongswanVersion, strongDucklingVersion}),
		tcpChecker: &tcpChecker{
			open: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemTcpChecker,
				Name:      "open_info",
				Help:      "Is TCP open is 1 otherwise 0",
			}, []string{"name", "address", "port"}),
			connectedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemTcpChecker,
				Name:      "connected_total",
				Help:      "Total number of times connection to TCP address:port was established",
			}, []string{"name", "address", "port"}),
			disconectedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemTcpChecker,
				Name:      "disconnected_total",
				Help:      "Total number of times connection to TCP address:port was lost",
			}, []string{"name", "address", "port"}),
		},
		ikeSA: ikeSA{
			establishedSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "established_seconds",
				Help:      "Number of seconds the SA has been established",
			}, []string{}),
			packetsIn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_in_total",
				Help:      "Total number of received packets",
			}, []string{childSAName}),
			packetsOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_out_total",
				Help:      "Total number of transmitted packets",
			}, []string{childSAName}),
			lastPacketInSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_in_silence_duration_seconds",
				Help:      "Duration of silences between packets in",
				Buckets:   prometheus.ExponentialBuckets(0.1, 10, 10),
			}, []string{}),
			lastPacketOutSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_out_silence_duration_seconds",
				Help:      "Duration of silences between packets out",
				Buckets:   prometheus.ExponentialBuckets(0.1, 10, 10),
			}, []string{}),
			bytesIn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "bytes_in_total",
				Help:      "Total number of received bytes",
			}, []string{childSAName}),
			bytesOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "bytes_out_total",
				Help:      "Total number of transmitted bytes",
			}, []string{childSAName}),
			installedSeconds: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "installs_total",
				Help:      "Total number of SA installs",
			}, []string{}),
			rekeySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "rekey_seconds",
				Help:      "Duration of each key session",
				Buckets:   prometheus.ExponentialBuckets(0.1, 10800, 10), // 3 hours
			}, []string{}),
			lifeTimeSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "lifetime_seconds",
				Help:      "Duration of each IKE session",
				Buckets:   prometheus.ExponentialBuckets(0.1, 10800, 10), // 3 hours
			}, []string{}),
			state: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "state_info",
				Help:      "Current state of the SA",
			}, []string{}),
			childSAState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "child_state_info",
				Help:      "Current state of the child SA",
			}, []string{}),
		},
	}

	err := register(r.registry,
		r.version,
		r.tcpChecker.open,
		r.tcpChecker.connectedTotal,
		r.tcpChecker.disconectedTotal,
		r.ikeSA.establishedSeconds,
		r.ikeSA.packetsIn,
		r.ikeSA.packetsOut,
		r.ikeSA.lastPacketInSeconds,
		r.ikeSA.lastPacketOutSeconds,
		r.ikeSA.bytesIn,
		r.ikeSA.bytesOut,
		r.ikeSA.installedSeconds,
		r.ikeSA.rekeySeconds,
		r.ikeSA.lifeTimeSeconds,
		r.ikeSA.state,
		r.ikeSA.childSAState,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func register(r prometheus.Registerer, collectors ...prometheus.Collector) error {
	for _, c := range collectors {
		err := r.Register(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PrometheusReporter) Info(strongDucklingVersion string) {
	p.version.WithLabelValues(strongDucklingVersion).Set(1)
}

func (r *tcpChecker) ReportPortCheck(report tcpchecker.Report) {
	if report.Open {
		r.open.WithLabelValues(report.Name, report.Address, fmt.Sprintf("%d", report.Port)).Set(1)
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.connectedTotal.WithLabelValues(report.Name, report.Address, fmt.Sprintf("%d", report.Port)).Add(1)
		}
	} else {
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.disconectedTotal.WithLabelValues(report.Name, report.Address, fmt.Sprintf("%d", report.Port)).Add(0)
		}
	}
	r.previousOpenState = &report.Open
}

func (p *PrometheusReporter) IKESAStatus(conn vici.IKEConf, sa *vici.IkeSa) {
	p.setEstablishedSeconds(sa)
	for name, child := range sa.ChildSAs {
		p.setGauge(p.ikeSA.packetsIn, child.PacketsIn, "packets in", name)
		p.setGauge(p.ikeSA.packetsOut, child.PacketsOut, "packets out", name)
		p.setGauge(p.ikeSA.bytesIn, child.BytesIn, "bytes in", name)
		p.setGauge(p.ikeSA.bytesOut, child.BytesOut, "bytes out", name)
	}
}

// setEstablishedSeconds sets gauge establishedSeconds if its value has
// decreased from the last call to it.
//
// The value is ever increasing as long as the IKE session is established so we
// should only mark the duration when it has reset, ie. a new connection is
// established.
func (p *PrometheusReporter) setEstablishedSeconds(sa *vici.IkeSa) {
	f, err := strconv.ParseFloat(sa.EstablishedSeconds, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert establishedSeconds '%s' to float64: %v", sa.EstablishedSeconds, err)
		return
	}
	// store the value for future reference when this call finishes
	defer func() {
		p.ikeSA.previousEstablishedSeconds = f
	}()
	if p.ikeSA.previousEstablishedSeconds > f {
		p.ikeSA.establishedSeconds.WithLabelValues().Set(p.ikeSA.previousEstablishedSeconds)
	}
}

func (p *PrometheusReporter) setGauge(g *prometheus.GaugeVec, value, name string, lbv ...string) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return
	}
	g.WithLabelValues(lbv...).Set(f)
}

func (p *PrometheusReporter) IKEConnectionConfiguration(name string, conf vici.IKEConf) {
}
