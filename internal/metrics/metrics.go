package metrics

import (
	"fmt"
	"net/http"

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
)

// metric label consts
const (
	strongswanVersion     = "strongswan_version"
	strongDucklingVersion = "version"
)

type PrometheusReporter struct {
	registry   prometheus.Registerer
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
	establishedSeconds   *prometheus.GaugeVec
	packetsIn            *prometheus.CounterVec
	packetsOut           *prometheus.CounterVec
	lastPacketInSeconds  *prometheus.HistogramVec
	lastPacketOutSeconds *prometheus.HistogramVec
	bytesIn              *prometheus.CounterVec
	bytesOut             *prometheus.CounterVec
	installedSeconds     *prometheus.CounterVec
	rekeySeconds         *prometheus.HistogramVec
	lifeTimeSeconds      *prometheus.HistogramVec
	state                *prometheus.GaugeVec
	childSAState         *prometheus.GaugeVec
}

func NewPrometheusReporter(reg prometheus.Registerer) (*PrometheusReporter, error) {
	r := PrometheusReporter{
		registry: reg,
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
				Subsystem: "ika_sa",
				Name:      "established_seconds",
				Help:      "Number of seconds the SA has been established",
			}, []string{}),
			packetsIn: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "packets_in_total",
				Help:      "Total number of received packets",
			}, []string{}),
			packetsOut: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "packets_out_total",
				Help:      "Total number of transmitted packets",
			}, []string{}),
			lastPacketInSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "packets_in_silence_duration_seconds",
				Help:      "Duration of silences between packets in",
			}, []string{}),
			lastPacketOutSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "packets_out_silence_duration_seconds",
				Help:      "Duration of silences between packets out",
			}, []string{}),
			bytesIn: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "bytes_in_total",
				Help:      "Total number of received bytes",
			}, []string{}),
			bytesOut: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "bytes_out_total",
				Help:      "Total number of transmitted bytes",
			}, []string{}),
			installedSeconds: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "installs_total",
				Help:      "Total number of SA installs",
			}, []string{}),
			rekeySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "rekey_seconds",
				Help:      "Duration of each key session",
			}, []string{}),
			lifeTimeSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "lifetime_seconds",
				Help:      "Duration of each IKE session",
			}, []string{}),
			state: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "state_info",
				Help:      "Current state of the SA",
			}, []string{}),
			childSAState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "ika_sa",
				Name:      "child_state_info",
				Help:      "Current state of the child SA",
			}, []string{}),
		},
	}

	r.registry.MustRegister(
		r.version,
		r.tcpChecker.open,
		r.tcpChecker.connectedTotal,
		r.tcpChecker.disconectedTotal)

	return &r, nil
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
}

func (p *PrometheusReporter) IKEConnectionConfiguration(name string, conf vici.IKEConf) {
}
