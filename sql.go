package prometheus_utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SqlMetrics interface {
	Inc(query, success string)
	WriteTiming(start time.Time, query, success string)
}

// sqlMetrics is a struct that allows to write metrics of count and latency of sql queries
type sqlMetrics struct {
	queries *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

var _ SqlMetrics = (*sqlMetrics)(nil)

func NewSqlMetrics(namespace, subsystem, service, host, dbName string) *sqlMetrics {
	queriesCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queries_count",
			Help: "How many queries processed.",
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
				"host":      host,
				"db":        dbName,
			},
		},
		[]string{"query", "success"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "queries_latency_milliseconds",
		Help: "How long it took to process the query.",
		ConstLabels: prometheus.Labels{
			"namespace": namespace,
			"subsystem": subsystem,
			"service":   service,
			"host":      host,
			"db":        dbName,
		},
		Buckets: []float64{200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200, 1500, 2000},
	},
		[]string{"query", "success"},
	)

	prometheus.MustRegister(queriesCollector, latencyCollector)

	return &sqlMetrics{
		queries: queriesCollector,
		latency: latencyCollector,
	}
}

// Inc increases the counter for the given "query" and "success" fields by 1
func (h *sqlMetrics) Inc(query, success string) {
	h.queries.WithLabelValues(query, success).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// for the given "query" and "success" fields
func (h *sqlMetrics) WriteTiming(startTime time.Time, query, success string) {
	h.latency.WithLabelValues(query, success).Observe(MillisecondsFromStart(startTime))
}
