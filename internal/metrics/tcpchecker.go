package metrics

import (
	"fmt"

	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSystemTcpChecker = "tcp_checker"
)

type tcpChecker struct {
	checks           *prometheus.CounterVec
	open             *prometheus.GaugeVec
	connectedTotal   *prometheus.CounterVec
	disconectedTotal *prometheus.CounterVec

	previousOpenState *bool
}

func newTcpChecker() *tcpChecker {
	return &tcpChecker{
		checks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystemTcpChecker,
			Name:      "checked_total",
			Help:      "Total number of times the connection has been checked",
		}, []string{"name", "address", "port", "open"}),
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
	}
}

func (tc *tcpChecker) getCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		tc.open,
		tc.checks,
		tc.connectedTotal,
		tc.disconectedTotal,
	}
}

func (r *tcpChecker) ReportPortCheck(report tcpchecker.Report) {
	labelValues := []string{report.Name, report.Address, fmt.Sprintf("%d", report.Port)}
	if report.Open {
		r.checks.WithLabelValues(append(labelValues, "true")...).Inc()
		r.open.WithLabelValues(labelValues...).Set(1)
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.connectedTotal.WithLabelValues(labelValues...).Add(1)
		}
	} else {
		r.checks.WithLabelValues(append(labelValues, "false")...).Inc()
		r.open.WithLabelValues(labelValues...).Set(0)
		if r.previousOpenState == nil || *r.previousOpenState != report.Open {
			r.disconectedTotal.WithLabelValues(labelValues...).Add(0)
		}
	}
	r.previousOpenState = &report.Open
}
