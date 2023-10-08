package prometheus_utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

// gauge is a struct that allows to add values
type dbGauge struct {
	gaugeVec *prometheus.GaugeVec
}

var _ Gauge = (*dbGauge)(nil)

func NewDBGauge(namespace, subsystem, service, host, dbName string) *dbGauge {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gauge",
		Help: "What is the value of the parameter.",
		ConstLabels: prometheus.Labels{
			"namespace": namespace,
			"subsystem": subsystem,
			"service":   service,
			"host":      host,
			"db":        dbName,
		},
	},
		[]string{"value_name"},
	)

	prometheus.MustRegister(gaugeVec)

	return &dbGauge{
		gaugeVec: gaugeVec,
	}
}

// Add function adds a given value to the gauge
func (g *dbGauge) Add(valueName string, value float64) {
	g.gaugeVec.WithLabelValues(valueName).Add(value)
}

// Set function sets a given value to the gauge
func (g *dbGauge) Set(valueName string, value float64) {
	g.gaugeVec.WithLabelValues(valueName).Set(value)
}
