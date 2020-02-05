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
	strongDucklingVersion = "version"
	childSAName           = "child_sa_name"
)

type Logger interface {
	Infof(string, ...interface{})
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
	previousValues map[string]float64

	establishedSeconds   *prometheus.GaugeVec
	packetsIn            *prometheus.GaugeVec
	packetsOut           *prometheus.GaugeVec
	lastPacketInSeconds  *prometheus.HistogramVec
	lastPacketOutSeconds *prometheus.HistogramVec
	bytesIn              *prometheus.GaugeVec
	bytesOut             *prometheus.GaugeVec
	installs             *prometheus.CounterVec
	rekeySeconds         *prometheus.HistogramVec
	lifeTimeSeconds      *prometheus.HistogramVec
	state                *prometheus.GaugeVec
	childSAState         *prometheus.GaugeVec
}

func NewPrometheusReporter(reg prometheus.Registerer, logger Logger) (*PrometheusReporter, error) {
	r := PrometheusReporter{
		registry: reg,
		logger:   logger,
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Version info of strong_duckling",
		}, []string{strongDucklingVersion}),
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
			previousValues: make(map[string]float64),
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
			installs: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "installs_total",
				Help:      "Total number of SA installs",
			}, []string{}),
			rekeySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "rekey_seconds",
				Help:      "Duration between re-keying",
				Buckets:   []float64{10, 30, 60, 120, 300, 480, 600},
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
		r.ikeSA.installs,
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
	if sa == nil {
		p.logger.Errorf("No SA for connecetion configuration: %#v", conn)
		return
	}
	p.setGaugeByMax(p.ikeSA.establishedSeconds, sa.EstablishedSeconds, "EstablishedSeconds")
	p.logger.Infof("prometheusReporter: IKESAStatus: IKE_SA state: %v", sa.State)
	for name, child := range sa.ChildSAs {
		p.logger.Infof("prometheusReporter: IKESAStatus: IKE_SA child state: %v", child.State)
		p.setCounterByMax(p.ikeSA.installs, child.InstallTimeSeconds, "InstallTimeSeconds")
		p.setGauge(p.ikeSA.packetsIn, child.PacketsIn, "PacketsIn", name)
		p.setGauge(p.ikeSA.packetsOut, child.PacketsOut, "PacketsOut", name)
		p.setGauge(p.ikeSA.bytesIn, child.BytesIn, "BytesIn", name)
		p.setGauge(p.ikeSA.bytesOut, child.BytesOut, "BytesOut", name)
		p.setHistogramByMax(p.ikeSA.lastPacketInSeconds, child.LastPacketInSeconds, "LastPacketInSeconds")
		p.setHistogramByMax(p.ikeSA.lastPacketOutSeconds, child.LastPacketOutSeconds, "LastPacketOutSeconds")
		p.setHistogramByMin(p.ikeSA.rekeySeconds, child.RekeyTimeSeconds, "RekeyTimeSeconds")
		p.setHistogramByMax(p.ikeSA.lifeTimeSeconds, child.LifeTimeSeconds, "LifeTimeSeconds")
		p.setRekeySeconds(conn, child)
	}
}

func (p *PrometheusReporter) setRekeySeconds(conn vici.IKEConf, child vici.ChildSA) {
	// RekeyTimeSeconds on the conn conf is the start value and on the child the
	// time left from this value. We want to track how long each rekey session
	// was, ie. the ellapsed time from max to when it increases again. This is
	// done by finding a min value on the child field and subtracting that from
	// the max value on the conf.
	minRekeyTimeSeconds, ok := p.minValue("RekeyTimeSeconds", child.RekeyTimeSeconds)
	if !ok {
		return
	}
	connRekeyTimeSeconds, err := strconv.ParseFloat(conn.RekeyTimeSeconds, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert RekeyTimeSeconds '%s' to float64: %v", conn.RekeyTimeSeconds, err)
		return
	}
	p.ikeSA.rekeySeconds.WithLabelValues().Observe(connRekeyTimeSeconds - minRekeyTimeSeconds)
}

func (p *PrometheusReporter) setGauge(g *prometheus.GaugeVec, value, name string, lbv ...string) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return
	}
	g.WithLabelValues(lbv...).Set(f)
}

func (p *PrometheusReporter) setCounterByMax(c *prometheus.CounterVec, value, name string) {
	// if this is the first time it is called it should be increased as well
	_, ok := p.ikeSA.previousValues[name]
	if !ok {
		c.WithLabelValues().Inc()
	}
	_, ok = p.maxValue(name, value)
	if !ok {
		return
	}
	c.WithLabelValues().Inc()
}

func (p *PrometheusReporter) setGaugeByMax(g *prometheus.GaugeVec, value, name string) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	g.WithLabelValues().Set(max)
}

func (p *PrometheusReporter) setHistogramByMax(h *prometheus.HistogramVec, value, name string) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues().Observe(max)
}

func (p *PrometheusReporter) setHistogramByMin(h *prometheus.HistogramVec, value, name string) {
	max, ok := p.minValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues().Observe(max)
}

// maxValue detects the max value of value. If max is detected the returned
// bool is true otherwise it returns the current value.
func (p *PrometheusReporter) maxValue(name, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return 0, false
	}
	previousValue, ok := p.ikeSA.previousValues[name]
	// store the value for future reference when this call finishes
	p.ikeSA.previousValues[name] = f
	if ok && previousValue > f {
		return previousValue, true
	}
	return f, false
}

// minValue detects the min value of value. If min is detected the returned
// bool is true otherwise it returns the current value.
func (p *PrometheusReporter) minValue(name, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return 0, false
	}
	previousValue, ok := p.ikeSA.previousValues[name]
	// store the value for future reference when this call finishes
	p.ikeSA.previousValues[name] = f
	if ok && previousValue < f {
		return previousValue, true
	}
	return f, false
}
