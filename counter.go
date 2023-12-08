package prometheus_utils

import "github.com/prometheus/client_golang/prometheus"

// counter counts the events grouped by specified labels
type counter struct {
	*prometheus.CounterVec
}

// NewCounter creates a new named counter for the app
// labels represent the actions which will be counted
func NewCounter(namespace, subsystem, service, counterName, subject string, labels ...string) *counter {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: counterName,
			Help: "Counts " + counterName,
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
				"subject":   subject,
			},
		},
		labels,
	)
	prometheus.MustRegister(c)
	return &counter{c}
}

// Inc increments the counter for the given label values by 1
func (c *counter) Inc(labelValues ...string) {
	c.WithLabelValues(labelValues...).Inc()
}

// Add adds the val to the counter with the given label values
func (c *counter) Add(val int64, labelValues ...string) {
	c.WithLabelValues(labelValues...).Add(float64(val))
}
