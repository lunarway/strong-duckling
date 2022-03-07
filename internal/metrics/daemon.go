package metrics

import (
	"time"

	"github.com/go-kit/kit/log"
	daemonpkg "github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSystemDaemon = "daemon"
)

type daemon struct {
	logger log.Logger

	started *prometheus.CounterVec
	stopped *prometheus.CounterVec
	skipped *prometheus.CounterVec
	ticked  *prometheus.CounterVec
}

func newDaemon() *daemon {
	return &daemon{
		started: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystemDaemon,
			Name:      "starts_total",
			Help:      "Total number of times started",
		}, []string{"name", "interval"}),
		stopped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystemDaemon,
			Name:      "stops_total",
			Help:      "Total number of times stopped",
		}, []string{"name"}),
		skipped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystemDaemon,
			Name:      "skips_total",
			Help:      "Total number of times tick was skipped",
		}, []string{"name"}),
		ticked: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystemDaemon,
			Name:      "ticks_total",
			Help:      "Total number of times tick was invoked",
		}, []string{"name"}),
	}
}

func (d *daemon) getCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		d.started,
		d.stopped,
		d.skipped,
		d.ticked,
	}
}

func (d *daemon) DefaultDaemonReporter(logger log.Logger, name string) *daemonpkg.Reporter {
	return &daemonpkg.Reporter{
		Started: func(duration time.Duration) {
			logger.With("state", "started").Infof("%s daemon started with interval %v", name, duration)
			d.started.WithLabelValues(name, duration.String()).Inc()
		},
		Stopped: func() {
			logger.With("state", "stopped").Infof("%s daemon stopped", name)
			d.stopped.WithLabelValues(name).Inc()
		},
		Skipped: func() {
			d.skipped.WithLabelValues(name).Inc()
		},
		Ticked: func() {
			d.ticked.WithLabelValues(name).Inc()
		},
	}
}
