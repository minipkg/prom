package prometheus_utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HttpClientMetric interface {
	Inc(method, code, path string)
	WriteTiming(startTime time.Time, method, code, path string)
}

// httpClientMetrics is a struct that allows to write metrics of count and latency of http requests
type httpClientMetric struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

var _ HttpClientMetric = (*httpClientMetric)(nil)

func NewHttpClientMetrics(namespace, subsystem, service, remoteService string) *httpClientMetric {
	reqsCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "client_reqs_count",
			Help: "How many HTTP requests processed",
			ConstLabels: prometheus.Labels{
				"namespace":      namespace,
				"subsystem":      subsystem,
				"service":        service,
				"remote_service": remoteService,
			},
		},
		[]string{"method", "status", "path"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "client_reqs_latency_milliseconds",
		Help: "How long it took to process the request",
		ConstLabels: prometheus.Labels{
			"namespace":      namespace,
			"subsystem":      subsystem,
			"service":        service,
			"remote_service": remoteService,
		},
		Buckets: []float64{200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200, 1500, 2000},
	},
		[]string{"method", "status", "path"},
	)

	prometheus.MustRegister(reqsCollector, latencyCollector)

	return &httpClientMetric{
		reqs:    reqsCollector,
		latency: latencyCollector,
	}
}

// Inc increases requests counter by one. method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) Inc(method, code, path string) {
	h.reqs.WithLabelValues(method, code, path).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) WriteTiming(startTime time.Time, method, code, path string) {
	h.latency.WithLabelValues(method, code, path).Observe(MillisecondsFromStart(startTime))
}
