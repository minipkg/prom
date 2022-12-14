package mp_prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// sqlMetrics is a struct that allows to write metrics of count and latency of sql queries
type sqlMetrics struct {
	queries *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func NewSqlMetrics(service, host string) *sqlMetrics {
	queriesCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "queries_count",
			Help:        "How many queries processed",
			ConstLabels: prometheus.Labels{"app": service, "host": host},
		},
		[]string{"query", "success"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "queries_latency",
		Help:        "How long it took to process the query",
		ConstLabels: prometheus.Labels{"app": service, "host": host},
		Buckets:     []float64{200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200, 1500, 2000},
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
	h.latency.WithLabelValues(query, success).Observe(timeFromStart(startTime))
}
