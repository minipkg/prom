// nolint
package mp_prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// mqMetrics is a struct that allows to write metrics of count and latency of msgs from broker
type mqMetrics struct {
	msgs *prometheus.CounterVec
	time *prometheus.HistogramVec
}

func NewMqMetrics(appName, host, subject string) *mqMetrics {
	msgsCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "msgs_count",
			Help:        "How many messages proceeded",
			ConstLabels: prometheus.Labels{"app": appName, "host": host, "subject": subject},
		},
		[]string{"status", "topic"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "msgs_latency",
		Help:        "How long it took to process the messages",
		ConstLabels: prometheus.Labels{"app": appName, "host": host, "subject": subject},
		Buckets:     []float64{200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200, 1500, 2000},
	},
		[]string{"status", "topic"},
	)

	prometheus.MustRegister(msgsCollector, latencyCollector)

	return &mqMetrics{
		msgs: msgsCollector,
		time: latencyCollector,
	}
}

// Inc increases the counter with given "status" and for the given "topic" labels by 1
func (h *mqMetrics) Inc(status, topic string) {
	h.msgs.WithLabelValues(status, topic).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// status, topic and path are label values for the "status" and "topic" fields
func (h *mqMetrics) WriteTiming(startTime time.Time, status, topic string) {
	h.time.WithLabelValues(status, topic).Observe(timeFromStart(startTime))
}
