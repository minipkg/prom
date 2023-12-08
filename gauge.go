package prometheus_utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Gauge interface {
	Add(valueName string, value float64)
	Set(valueName string, value float64)
}

// gauge is a struct that allows to add values
type gauge struct {
	gaugeVec *prometheus.GaugeVec
}

var _ Gauge = (*gauge)(nil)

func NewGauge(namespace, subsystem, service, gaugeName, subject string, labels ...string) *gauge {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: gaugeName,
			Help: "What is the value of " + gaugeName,
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
				"subject":   subject,
			},
		},
		labels,
	)

	prometheus.MustRegister(gaugeVec)

	return &gauge{
		gaugeVec: gaugeVec,
	}
}

// Add function adds a given value to the gauge
func (g *gauge) Add(valueName string, value float64) {
	g.gaugeVec.WithLabelValues(valueName).Add(value)
}

// Set function sets a given value to the gauge
func (g *gauge) Set(valueName string, value float64) {
	g.gaugeVec.WithLabelValues(valueName).Set(value)
}
