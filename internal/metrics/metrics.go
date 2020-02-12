package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	subSystemDaemon     = "daemon"
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
	Daemon     *Daemon
}

func (pr *PrometheusReporter) TcpChecker() tcpchecker.Reporter {
	return pr.tcpChecker
}

type tcpChecker struct {
	checks           *prometheus.CounterVec
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

type ikeSALabels struct {
	name, localPeerIP, remotePeerIP string
}

func (i ikeSALabels) names() []string {
	return []string{"ike_sa_name", "local_peer_ip", "remote_peer_ip"}
}

func (i ikeSALabels) values() []string {
	return []string{i.name, i.localPeerIP, i.remotePeerIP}
}

type childSALabels struct {
	ikeSALabels
	localIPRange, remoteIPRange, childSAName string
}

func (c childSALabels) names() []string {
	return append(c.ikeSALabels.names(), "local_ip_range", "remote_ip_range", "child_sa_name")
}

func (c childSALabels) values() []string {
	return append(c.ikeSALabels.values(), c.localIPRange, c.remoteIPRange, c.childSAName)
}

type Daemon struct {
	Started *prometheus.CounterVec
	Stopped *prometheus.CounterVec
	Ticked  *prometheus.CounterVec
	Skipped *prometheus.CounterVec
}

func NewPrometheusReporter(reg prometheus.Registerer, logger Logger) (*PrometheusReporter, error) {
	r := PrometheusReporter{
		registry: reg,
		logger:   logger,
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Version info of strong_duckling",
		}, []string{"version"}),
		tcpChecker: &tcpChecker{
			checks: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemTcpChecker,
				Name:      "checked_total",
				Help:      "Total number of times the connection has been checked",
			}, []string{"name", "address", "port"}),
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
			}, ikeSALabels{}.names()),
			packetsIn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_in_total",
				Help:      "Total number of received packets",
			}, childSALabels{}.names()),
			packetsOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_out_total",
				Help:      "Total number of transmitted packets",
			}, childSALabels{}.names()),
			lastPacketInSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_in_silence_duration_seconds",
				Help:      "Duration of silences between packets in",
				Buckets:   prometheus.ExponentialBuckets(15, 2, 14),
			}, childSALabels{}.names()),
			lastPacketOutSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "packets_out_silence_duration_seconds",
				Help:      "Duration of silences between packets out",
				Buckets:   prometheus.ExponentialBuckets(15, 2, 14),
			}, childSALabels{}.names()),
			bytesIn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "bytes_in_total",
				Help:      "Total number of received bytes",
			}, childSALabels{}.names()),
			bytesOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "bytes_out_total",
				Help:      "Total number of transmitted bytes",
			}, childSALabels{}.names()),
			installs: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "installs_total",
				Help:      "Total number of SA installs",
			}, childSALabels{}.names()),
			rekeySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "rekey_seconds",
				Help:      "Duration between re-keying",
				Buckets:   prometheus.ExponentialBuckets(15, 2, 12),
			}, childSALabels{}.names()),
			lifeTimeSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "lifetime_seconds",
				Help:      "Duration of each IKE session",
				Buckets:   prometheus.ExponentialBuckets(15, 2, 14),
			}, childSALabels{}.names()),
			state: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "state_info",
				Help:      "Current state of the SA",
			}, ikeSALabels{}.names()),
			childSAState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystemIKE,
				Name:      "child_state_info",
				Help:      "Current state of the child SA",
			}, childSALabels{}.names()),
		},
		Daemon: &Daemon{
			Started: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemDaemon,
				Name:      "starts_total",
				Help:      "Total number of times started",
			}, []string{"name", "interval"}),
			Stopped: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemDaemon,
				Name:      "stops_total",
				Help:      "Total number of times stopped",
			}, []string{"name"}),
			Skipped: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemDaemon,
				Name:      "skips_total",
				Help:      "Total number of times tick was skipped",
			}, []string{"name"}),
			Ticked: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystemDaemon,
				Name:      "ticks_total",
				Help:      "Total number of times tick was invoked",
			}, []string{"name"}),
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
		r.Daemon.Started,
		r.Daemon.Stopped,
		r.Daemon.Skipped,
		r.Daemon.Ticked,
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
	labelValues := []string{report.Name, report.Address, fmt.Sprintf("%d", report.Port)}
	r.checks.WithLabelValues(labelValues...).Inc()
	if report.Open {
		r.open.WithLabelValues(labelValues...).Set(1)
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.connectedTotal.WithLabelValues(labelValues...).Add(1)
		}
	} else {
		r.open.WithLabelValues(labelValues...).Set(0)
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.disconectedTotal.WithLabelValues(labelValues...).Add(0)
		}
	}
	r.previousOpenState = &report.Open
}

func (p *PrometheusReporter) IKESAStatus(ikeName string, conn vici.IKEConf, sa *vici.IkeSa) {
	if sa == nil {
		p.logger.Errorf("No SA for connecetion configuration: %#v", conn)
		return
	}
	ikeSALabels := ikeSALabels{
		name:         ikeName,
		localPeerIP:  sa.LocalHost,
		remotePeerIP: sa.RemoteHost,
	}
	p.setGaugeByMax(p.ikeSA.establishedSeconds, sa.EstablishedSeconds, "EstablishedSeconds", ikeSALabels)
	p.logger.Infof("prometheusReporter: IKESAStatus: IKE_SA state: %v", sa.State)
	for _, child := range sa.ChildSAs {
		labels := childSALabels{
			ikeSALabels:   ikeSALabels,
			childSAName:   child.Name,
			localIPRange:  strings.Join(child.LocalTrafficSelectors, ","),
			remoteIPRange: strings.Join(child.RemoteTrafficSelectors, ","),
		}
		p.logger.Infof("prometheusReporter: IKESAStatus: IKE_SA child state: %v", child.State)
		p.setCounterByMax(p.ikeSA.installs, child.InstallTimeSeconds, "InstallTimeSeconds", labels)
		p.setGauge(p.ikeSA.packetsIn, child.PacketsIn, "PacketsIn", labels)
		p.setGauge(p.ikeSA.packetsOut, child.PacketsOut, "PacketsOut", labels)
		p.setGauge(p.ikeSA.bytesIn, child.BytesIn, "BytesIn", labels)
		p.setGauge(p.ikeSA.bytesOut, child.BytesOut, "BytesOut", labels)
		p.setHistogramByMax(p.ikeSA.lastPacketInSeconds, child.LastPacketInSeconds, "LastPacketInSeconds", labels)
		p.setHistogramByMax(p.ikeSA.lastPacketOutSeconds, child.LastPacketOutSeconds, "LastPacketOutSeconds", labels)
		p.setHistogramByMin(p.ikeSA.rekeySeconds, child.RekeyTimeSeconds, "RekeyTimeSeconds", labels)
		p.setHistogramByMax(p.ikeSA.lifeTimeSeconds, child.LifeTimeSeconds, "LifeTimeSeconds", labels)
		p.setRekeySeconds(conn, child, labels)
	}
}

func (p *PrometheusReporter) setRekeySeconds(conn vici.IKEConf, child vici.ChildSA, labels childSALabels) {
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
	p.ikeSA.rekeySeconds.WithLabelValues(labels.values()...).Observe(connRekeyTimeSeconds - minRekeyTimeSeconds)
}

func (p *PrometheusReporter) setGauge(g *prometheus.GaugeVec, value, name string, labels childSALabels) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return
	}
	g.WithLabelValues(labels.values()...).Set(f)
}

func (p *PrometheusReporter) setCounterByMax(c *prometheus.CounterVec, value, name string, labels childSALabels) {
	// if this is the first time it is called it should be increased as well
	_, ok := p.ikeSA.previousValues[name]
	if !ok {
		c.WithLabelValues(labels.values()...).Inc()
	}
	_, ok = p.maxValue(name, value)
	if !ok {
		return
	}
	c.WithLabelValues(labels.values()...).Inc()
}

func (p *PrometheusReporter) setGaugeByMax(g *prometheus.GaugeVec, value, name string, labels ikeSALabels) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	g.WithLabelValues(labels.values()...).Set(max)
}

func (p *PrometheusReporter) setHistogramByMax(h *prometheus.HistogramVec, value, name string, labels childSALabels) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues(labels.values()...).Observe(max)
}

func (p *PrometheusReporter) setHistogramByMin(h *prometheus.HistogramVec, value, name string, labels childSALabels) {
	min, ok := p.minValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues(labels.values()...).Observe(min)
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
