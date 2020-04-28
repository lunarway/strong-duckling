package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func newHelper(logger log.Logger) *helper {
	return &helper{
		previousValues: make(map[string]float64),
		logger:         logger,
	}
}

type helper struct {
	previousValues map[string]float64
	logger         log.Logger
}

func (p *helper) setGauge(g *prometheus.GaugeVec, value, name string, labels childSALabels) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return
	}
	g.WithLabelValues(labels.values()...).Set(f)
}

func (p *helper) setCounterByMax(c *prometheus.CounterVec, value, name string, labels childSALabels) {
	// if this is the first time it is called it should be increased as well
	_, ok := p.previousValues[name]
	if !ok {
		c.WithLabelValues(labels.values()...).Inc()
	}
	_, ok = p.maxValue(name, value)
	if !ok {
		return
	}
	c.WithLabelValues(labels.values()...).Inc()
}

func (p *helper) setGaugeByMax(g *prometheus.GaugeVec, value, name string, labels ikeSALabels) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	g.WithLabelValues(labels.values()...).Set(max)
}

func (p *helper) setHistogramByMax(h *prometheus.HistogramVec, value, name string, labels childSALabels) {
	max, ok := p.maxValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues(labels.values()...).Observe(max)
}

func (p *helper) setHistogramByMin(h *prometheus.HistogramVec, value, name string, labels childSALabels) {
	min, ok := p.minValue(name, value)
	if !ok {
		return
	}
	h.WithLabelValues(labels.values()...).Observe(min)
}

// maxValue detects the max value of value. If max is detected the returned
// bool is true otherwise it returns the current value.
func (p *helper) maxValue(name, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return 0, false
	}
	previousValue, ok := p.previousValues[name]
	// store the value for future reference when this call finishes
	p.previousValues[name] = f
	if ok && previousValue > f {
		return previousValue, true
	}
	return f, false
}

// minValue detects the min value of value. If min is detected the returned
// bool is true otherwise it returns the current value.
func (p *helper) minValue(name, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Errorf("metrics: failed to convert %s '%s' to float64: %v", name, value, err)
		return 0, false
	}
	previousValue, ok := p.previousValues[name]
	// store the value for future reference when this call finishes
	p.previousValues[name] = f
	if ok && previousValue < f {
		return previousValue, true
	}
	return f, false
}
