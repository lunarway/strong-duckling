package metrics

import (
	"strconv"
	"strings"

	"github.com/lunarway/strong-duckling/internal/strongswan"
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	subSystemIKE = "ike_sa"
)

type ikeSA struct {
	logger zap.Logger
	helper *helper

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

func newIkeSA(logger zap.Logger) *ikeSA {
	return &ikeSA{
		logger: logger,
		helper: newHelper(logger),
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
	}
}

func (i *ikeSA) getCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		i.establishedSeconds,
		i.packetsIn,
		i.packetsOut,
		i.lastPacketInSeconds,
		i.lastPacketOutSeconds,
		i.bytesIn,
		i.bytesOut,
		i.installs,
		i.rekeySeconds,
		i.lifeTimeSeconds,
		i.state,
		i.childSAState,
	}
}

func (p *ikeSA) IKESAStatus(ikeSAStatus strongswan.IKESAStatus) {
	if ikeSAStatus.State == nil {
		p.logger.Sugar().Errorf("No SA for connection configuration: %#v", ikeSAStatus.Configuration)
		return
	}
	ikeSALabels := ikeSALabels{
		name:         ikeSAStatus.Name,
		localPeerIP:  ikeSAStatus.State.LocalHost,
		remotePeerIP: ikeSAStatus.State.RemoteHost,
	}
	p.helper.setGaugeByMax(p.establishedSeconds, ikeSAStatus.State.EstablishedSeconds, "EstablishedSeconds", ikeSALabels)
	p.logger.Sugar().Infof("prometheusReporter: IKESAStatus: IKE_SA state: %v", ikeSAStatus.State.State)
	for _, child := range ikeSAStatus.State.ChildSAs {
		labels := childSALabels{
			ikeSALabels:   ikeSALabels,
			childSAName:   child.Name,
			localIPRange:  strings.Join(child.LocalTrafficSelectors, ","),
			remoteIPRange: strings.Join(child.RemoteTrafficSelectors, ","),
		}
		p.logger.Sugar().Infof("prometheusReporter: IKESAStatus: IKE_SA child state: %v", child.State)
		p.helper.setCounterByMax(p.installs, child.InstallTimeSeconds, "InstallTimeSeconds", labels)
		p.helper.setGauge(p.packetsIn, child.PacketsIn, "PacketsIn", labels)
		p.helper.setGauge(p.packetsOut, child.PacketsOut, "PacketsOut", labels)
		p.helper.setGauge(p.bytesIn, child.BytesIn, "BytesIn", labels)
		p.helper.setGauge(p.bytesOut, child.BytesOut, "BytesOut", labels)
		p.helper.setHistogramByMax(p.lastPacketInSeconds, child.LastPacketInSeconds, "LastPacketInSeconds", labels)
		p.helper.setHistogramByMax(p.lastPacketOutSeconds, child.LastPacketOutSeconds, "LastPacketOutSeconds", labels)
		p.helper.setHistogramByMin(p.rekeySeconds, child.RekeyTimeSeconds, "RekeyTimeSeconds", labels)
		p.helper.setHistogramByMax(p.lifeTimeSeconds, child.LifeTimeSeconds, "LifeTimeSeconds", labels)
		p.setRekeySeconds(ikeSAStatus.Configuration, child, labels)
	}
}

func (p *ikeSA) setRekeySeconds(conn vici.IKEConf, child vici.ChildSA, labels childSALabels) {
	// RekeyTimeSeconds on the conn conf is the start value and on the child the
	// time left from this value. We want to track how long each rekey session
	// was, ie. the ellapsed time from max to when it increases again. This is
	// done by finding a min value on the child field and subtracting that from
	// the max value on the conf.
	minRekeyTimeSeconds, ok := p.helper.minValue("RekeyTimeSeconds", child.RekeyTimeSeconds)
	if !ok {
		return
	}
	connRekeyTimeSeconds, err := strconv.ParseFloat(conn.RekeyTimeSeconds, 64)
	if err != nil {
		p.logger.Sugar().Errorf("metrics: failed to convert RekeyTimeSeconds '%s' to float64: %v", conn.RekeyTimeSeconds, err)
		return
	}
	p.rekeySeconds.WithLabelValues(labels.values()...).Observe(connRekeyTimeSeconds - minRekeyTimeSeconds)
}
