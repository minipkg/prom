package prometheus_utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type HttpServerMetric interface {
	Inc(method, code, path, client string)
	WriteTiming(startTime time.Time, method, code, path, client string)
}

// httpServerMetrics is a struct that allows to write metrics of count and latency of http requests
type httpServerMetric struct {
	cuttingPathOpts *CuttingPathOpts
	reqs            *prometheus.CounterVec
	latency         *prometheus.HistogramVec
}

type CuttingPathOpts struct {
	isNeedToRemoveIDsInPath   bool
	isNeedToRemoveQueryInPath bool
	boundaries4CuttingPath    *[2]uint
}

const (
	AuthClientKey = "http.client"
)

var _ HttpServerMetric = (*httpServerMetric)(nil)

func NewHttpServerMetrics(namespace, subsystem, service string) *httpServerMetric {
	reqsCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reqs_count",
			Help: "How many HTTP requests processed",
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
			},
		},
		[]string{"method", "status", "path", "client"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "reqs_latency_milliseconds",
		Help: "How long it took to process the request",
		ConstLabels: prometheus.Labels{
			"namespace": namespace,
			"subsystem": subsystem,
			"service":   service,
		},
		Buckets: []float64{5, 10, 20, 30, 50, 70, 100, 150, 200, 300, 500, 1000},
	},
		[]string{"method", "status", "path", "client"},
	)

	prometheus.MustRegister(reqsCollector, latencyCollector)

	return &httpServerMetric{
		reqs:    reqsCollector,
		latency: latencyCollector,
	}
}

func (h *httpServerMetric) SetCuttingPathOpts(cuttingPathOpts *CuttingPathOpts) *httpServerMetric {
	h.cuttingPathOpts = cuttingPathOpts
	return h
}

func (h *httpServerMetric) checkAndCutPath(path string) string {
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

// Inc increases requests counter by one.
//
//	method, code, path and client are label values for "method", "status", "path" and "client" fields
func (h *httpServerMetric) Inc(method, code, path, client string) {
	h.reqs.WithLabelValues(method, code, h.checkAndCutPath(path), client).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// method, code, path and client are label values for "method", "status", "path" and "client" fields
func (h *httpServerMetric) WriteTiming(startTime time.Time, method, code, path, client string) {
	h.latency.WithLabelValues(method, code, h.checkAndCutPath(path), client).Observe(MillisecondsFromStart(startTime))
}

// Handler with metrics for "github.com/fasthttp/router"
func GetFasthttpHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
}

// Middleware with metrics for "github.com/fasthttp/router"
func (m *httpServerMetric) FasthttpRouterMetricsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		now := time.Now()

		next(ctx)

		client := ""
		if s, ok := ctx.UserValue(AuthClientKey).(string); ok {
			client = s
		}

		status := strconv.Itoa(ctx.Response.StatusCode())
		path := string(ctx.Path())
		method := string(ctx.Method())

		m.Inc(method, status, path, client)
		m.WriteTiming(now, method, status, path, client)
	}
}

// Handler with metrics for "github.com/qiangxue/fasthttp-routing"
func GetFasthttpRoutingHandler() routing.Handler {
	return func(rctx *routing.Context) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(rctx.RequestCtx)
		return nil
	}
}

// Middleware with metrics for "github.com/qiangxue/fasthttp-routing"
func (m *httpServerMetric) FasthttpRoutingMetricsMiddleware(rctx *routing.Context) error {
	now := time.Now()

	rctx.Next()

	client := ""
	if s, ok := rctx.UserValue(AuthClientKey).(string); ok {
		client = s
	}

	status := strconv.Itoa(rctx.Response.StatusCode())
	path := string(rctx.Path())
	method := string(rctx.Method())

	m.Inc(method, status, path, client)
	m.WriteTiming(now, method, status, path, client)

	return nil
}
