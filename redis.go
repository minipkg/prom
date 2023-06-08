package prometheus_utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// redisMetrics is a struct that allows to write metrics of count and latency of redis queries
type redisMetrics struct {
	queries *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func NewRedisMetrics(service string) *redisMetrics {
	queriesCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "redis_queries_count",
			Help:        "How many queries processed",
			ConstLabels: prometheus.Labels{"app": service},
		},
		[]string{"query", "success"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "redis_queries_latency_seconds",
		Help:        "How long it took to process the query",
		ConstLabels: prometheus.Labels{"app": service},
		Buckets:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 20},
	},
		[]string{"query", "success"},
	)

	prometheus.MustRegister(queriesCollector, latencyCollector)

	return &redisMetrics{
		queries: queriesCollector,
		latency: latencyCollector,
	}
}

// Inc increases the counter for the given "query" and "success" fields by 1
func (h *redisMetrics) Inc(query, success string) {
	h.queries.WithLabelValues(query, success).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// for the given "query" and "success" fields
func (h *redisMetrics) WriteTiming(startTime time.Time, query, success string) {
	h.latency.WithLabelValues(query, success).Observe(MillisecondsFromStart(startTime))
}
