package prometheus_utils

import (
	"regexp"
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
	cuttingPathOpts *CuttingPathOpts
	reqs            *prometheus.CounterVec
	latency         *prometheus.HistogramVec
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

func (h *httpClientMetric) SetCuttingPathOpts(cuttingPathOpts *CuttingPathOpts) *httpClientMetric {
	h.cuttingPathOpts = cuttingPathOpts
	return h
}

func (h *httpClientMetric) checkAndCutPath(path string) string {
	if h.cuttingPathOpts == nil {
		return path
	}

	if h.cuttingPathOpts.isNeedToRemoveQueryInPath {
		path = strings.Split(path, "?")[0]
	}

	if h.cuttingPathOpts.boundaries4CuttingPath != nil {
		sl := strings.Split(path, "/")
		min := int(h.cuttingPathOpts.boundaries4CuttingPath[0])
		if min >= len(sl) {
			min = len(sl) - 1
		}
		max := int(h.cuttingPathOpts.boundaries4CuttingPath[1])
		if max > len(sl) {
			max = len(sl)
		}
		path = strings.Join(sl[min:max], "/")
	}

	// remove ids from path
	if h.cuttingPathOpts.isNeedToRemoveQueryInPath {
		uintID := regexp.MustCompile("^[\\d,]+$")
		sl := strings.Split(path, "/")
		nsl := make([]string, 0, len(sl))
		for _, s := range sl {
			if !uintID.MatchString(s) {
				nsl = append(nsl, s)
			}
		}
		path = strings.Join(sl, "/")
	}

	return path
}

// Inc increases requests counter by one. method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) Inc(method, code, path string) {
	h.reqs.WithLabelValues(method, code, h.checkAndCutPath(path)).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// method, code and path are label values for "method", "status" and "path" fields
func (h *httpClientMetric) WriteTiming(startTime time.Time, method, code, path string) {
	h.latency.WithLabelValues(method, code, h.checkAndCutPath(path)).Observe(MillisecondsFromStart(startTime))
}
