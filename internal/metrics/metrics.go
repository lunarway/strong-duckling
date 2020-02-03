package metrics

import (
	"fmt"
	"net/http"

	"github.com/lunarway/strong-duckling/internal/tcpchecker"
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

const (
	strongswanVersion     = "strongswan_version"
	strongDucklingVersion = "version"
)

type PrometheusReporter struct {
	registry   prometheus.Registerer
	version    *prometheus.GaugeVec
	tcpChecker *tcpChecker
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
