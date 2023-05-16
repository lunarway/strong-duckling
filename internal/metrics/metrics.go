package metrics

import (
	"net/http"

	daemonpkg "github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/lunarway/strong-duckling/internal/strongswan"
	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func Register(serveMux *http.ServeMux) {
	serveMux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(rw, r)
		}),
	))
}

const (
	namespace = "strong_duckling"
)

type PrometheusReporter struct {
	registry prometheus.Registerer
	logger   zap.Logger

	version    *prometheus.GaugeVec
	tcpChecker *tcpChecker
	ikeSA      *ikeSA
	daemon     *daemon
}

func (pr *PrometheusReporter) TcpChecker() tcpchecker.Reporter {
	return pr.tcpChecker
}

func (pr *PrometheusReporter) StrongSwan() strongswan.IKESAStatusReceiver {
	return pr.ikeSA
}

func (pr *PrometheusReporter) Daemon(logger zap.Logger, name string) *daemonpkg.Reporter {
	return pr.daemon.DefaultDaemonReporter(logger, name)
}

func NewPrometheusReporter(reg prometheus.Registerer, logger zap.Logger) (*PrometheusReporter, error) {
	r := PrometheusReporter{
		registry: reg,
		logger:   logger,
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Version info of strong_duckling",
		}, []string{"version"}),
		tcpChecker: newTcpChecker(),
		ikeSA:      newIkeSA(logger),
		daemon:     newDaemon(),
	}

	collectors := []prometheus.Collector{
		r.version,
	}

	collectors = append(collectors, r.tcpChecker.getCollectors()...)
	collectors = append(collectors, r.ikeSA.getCollectors()...)
	collectors = append(collectors, r.daemon.getCollectors()...)

	err := register(r.registry, collectors...)
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
