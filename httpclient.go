package prometheus_utils

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HttpClientMetric interface {
	Inc(method, code, path string)
	WriteTiming(startTime time.Time, method, code, path string)
}

// httpClientMetrics is a struct that allows to write metrics of count and latency of http requests
type httpClientMetric struct {
	cuttingPathOpts CuttingPathOpts
	reqs            *prometheus.CounterVec
	latency         *prometheus.HistogramVec
}

type CuttingPathOpts struct {
	isNeedToRemoveQueryInPath bool
	boundaries4CuttingPath    *[2]uint
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

func (h *httpClientMetric) SetCuttingPathOpts(isNeedToRemoveQueryInPath bool, boundaries4CuttingPath *[2]uint) *httpClientMetric {
	h.cuttingPathOpts.isNeedToRemoveQueryInPath = isNeedToRemoveQueryInPath
	h.cuttingPathOpts.boundaries4CuttingPath = boundaries4CuttingPath
	return h
}

func (h *httpClientMetric) checkAndCutPath(path *string) {
	if h.cuttingPathOpts.isNeedToRemoveQueryInPath {
		*path = strings.Split(*path, "?")[0]
	}

	if h.cuttingPathOpts.boundaries4CuttingPath != nil {
		sl := strings.Split(*path, "/")[int(h.cuttingPathOpts.boundaries4CuttingPath[0]):int(h.cuttingPathOpts.boundaries4CuttingPath[1])]
		*path = strings.Join(sl, "/")
	}
	return
}

// Inc increases requests counter by one. method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) Inc(method, code, path string) {
	h.checkAndCutPath(&path)
	h.reqs.WithLabelValues(method, code, path).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) WriteTiming(startTime time.Time, method, code, path string) {
	h.checkAndCutPath(&path)
	h.latency.WithLabelValues(method, code, path).Observe(MillisecondsFromStart(startTime))
}
